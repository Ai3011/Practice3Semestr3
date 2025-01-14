package operations

import (
	"API/orderlogic"
	"API/utilities"
	"API/zapros"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// баланс
type BalanceResponse struct {
	Lot_id   int     `json:"lot_id"`
	Quantity float64 `json:"quantity"`
}

func HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	// Получаем ключ пользователя из заголовка запроса
	userKey := r.Header.Get("X-USER-KEY")

	// Проверяем наличие заголовка X-USER-KEY
	if userKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		fmt.Println("Empty user key")
		return
	}

	// SQL-запрос для получения идентификатора пользователя по его ключу
	var getUserQuery string = "SELECT user.user_id, user.key FROM user WHERE user.key = '" + userKey + "'"

	// Выполняем запрос к базе данных
	userData, err := zapros.RquestDataBase(getUserQuery)
	if err != nil {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		fmt.Println("User unauthorized")
		return
	}

	// Разделяем ответ для извлечения идентификатора пользователя
	userDataFields := strings.Split(userData, " ")
	if len(userDataFields) < 1 {
		http.Error(w, "User not found", http.StatusUnauthorized)
		fmt.Println("User not found")
		return
	}
	userID := userDataFields[0]

	// SQL-запрос для получения баланса пользователя по его идентификатору
	var getBalanceQuery string = "SELECT user_lot.lot_id, user_lot.quantity FROM user_lot WHERE user_lot.user_id = '" + userID + "'"

	// Выполняем запрос к базе данных для получения данных о балансе
	balanceResponse, err2 := zapros.RquestDataBase(getBalanceQuery)
	if err2 != nil {
		http.Error(w, "Failed to retrieve balance", http.StatusInternalServerError)
		fmt.Println("Failed to retrieve balance", err2)
		return
	}

	// Разделяем ответ базы данных на строки
	balanceRows := strings.Split(strings.TrimSpace(balanceResponse), "\n")

	// Массив для хранения балансов пользователя
	var balances []BalanceResponse

	// Обрабатываем каждую строку ответа
	for _, balanceRow := range balanceRows {
		// Разделяем строку на отдельные поля
		fields := strings.Split(balanceRow, " ")
		if len(fields) < 2 {
			continue // Пропускаем строки с недостаточным количеством полей
		}

		// Преобразуем данные из строки в нужный формат
		lotID, _ := strconv.Atoi(strings.TrimSpace(fields[0]))              // Идентификатор лота
		quantity, _ := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64) // Количество

		// Создаем структуру баланса
		balance := BalanceResponse{
			Lot_id:   lotID,
			Quantity: quantity,
		}

		// Добавляем структуру в массив балансов
		balances = append(balances, balance)
	}

	// Устанавливаем заголовок ответа и отправляем данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("Balance gived for user: ", userID)
	json.NewEncoder(w).Encode(balances)
}

// лоты
type LotResponse struct {
	Lot_id int    `json:"lot_id"`
	Name   string `json:"name"`
}

// Получение информации о лотах
func HandleGetLot(w http.ResponseWriter, r *http.Request) {
	// SQL-запрос для получения всех лотов
	var getLotsQuery string = "SELECT * FROM lot"

	// Выполняем запрос к базе данных
	dbResponse, err := zapros.RquestDataBase(getLotsQuery)
	if err != nil {
		fmt.Printf("Error getting: %v\n", err)
		return // Если ошибка, выходим из функции
	}

	// Разделяем ответ базы данных на строки
	dbRows := strings.Split(strings.TrimSpace(dbResponse), "\n")

	// Массив для хранения данных о лотах
	var lotResponses []LotResponse

	// Обрабатываем каждую строку ответа
	for _, dbRow := range dbRows {
		// Разделяем строку на отдельные поля
		fields := strings.Split(dbRow, " ")
		if len(fields) < 2 {
			continue // Пропускаем строки с недостаточным количеством полей
		}

		// Преобразуем данные из строки в нужный формат
		lotID, _ := strconv.Atoi(strings.TrimSpace(fields[0])) // Идентификатор лота
		lotName := strings.TrimSpace(fields[1])                // Название лота

		// Создаем структуру ответа
		lotResponse := LotResponse{
			Lot_id: lotID,
			Name:   lotName,
		}

		// Добавляем структуру в массив ответов
		lotResponses = append(lotResponses, lotResponse)
	}

	// Устанавливаем заголовок ответа и отправляем данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lotResponses)
}

// ордеры
type CreateOrderRequestStruct struct {
	PairID   int     `json:"pair_id"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Type     string  `json:"type"`
}

// Структура ответа при создании ордера
type CreateOrderResponseStruct struct {
	OrderID int `json:"order_id"`
}

type GetOrderResponseStruct struct {
	OrderID  int     `json:"order_id"`
	UserID   int     `json:"user_id"`
	PairID   int     `json:"lot_id"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Type     string  `json:"type"`
	Closed   string  `json:"closed"`
}

// Структура запроса на удаление ордера
type DeleteOrderStruct struct {
	OrderID int `json:"order_id"`
}

func CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Получаем ключ пользователя из заголовка запроса
	userKey := r.Header.Get("X-USER-KEY")

	// Проверка наличия заголовка X-USER-KEY, проверить есть ли такой пользователь
	if userKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		fmt.Println("Empty user key")
		return
	}

	// Парсинг JSON-запроса
	var req CreateOrderRequestStruct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		fmt.Println("Invalid JSON")
		return
	}

	// Проверка наличия ключа пользователя в БД
	var reqUserID string = "SELECT user.user_id FROM user WHERE user.key = '" + userKey + "'"
	userID, err := zapros.RquestDataBase(reqUserID)
	if err != nil || userID == "" {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		fmt.Println("User unauthorized")
		return
	}
	userID = userID[:len(userID)-2] // Убираем лишние символы из строки

	// Проверка наличия пары в БД
	var reqPairID string = "SELECT pair.pair_id FROM pair WHERE pair.pair_id = '" + strconv.Itoa(req.PairID) + "'"
	pairID, err1 := zapros.RquestDataBase(reqPairID)
	if err1 != nil || pairID == "" {
		http.Error(w, "Pair not found", http.StatusNotFound)
		fmt.Println("Pair not found")
		return
	}

	// Списание средств со счета пользователя
	payErr := orderlogic.PayByOrder(userID, req.PairID, req.Quantity, req.Price, req.Type, true)
	if payErr != nil {
		http.Error(w, "Not enough funds", http.StatusPaymentRequired)
		fmt.Println("Not enough funds", payErr)
		return
	}

	// Поиск подходящего ордера на покупку/продажу, если нашелся, начисляем новые средства
	newQuant, searchError := orderlogic.SearchOrder(userID, req.PairID, req.Type, req.Quantity, req.Price, req.Type)
	if searchError != nil {
		http.Error(w, "Not enough orders", http.StatusNotFound)
		fmt.Println("Not enough orders")
		return
	}

	// Создаем ордер
	status := ""
	if newQuant == 0 {
		status = "close"
		newQuant = req.Quantity
	} else if newQuant != req.Quantity {
		// Вносим в базу уже закрытый ордер (точнее его часть)
		var closeOrderQuery string = "INSERT INTO order VALUES ('" + userID + "', '" + strconv.Itoa(req.PairID) + "', '" + strconv.FormatFloat(req.Quantity, 'f', -1, 64) + "', '" + strconv.FormatFloat(req.Price, 'f', -1, 64) + "', '" + req.Type + "', 'close')"
		_, err := zapros.RquestDataBase(closeOrderQuery)
		if err != nil {
			fmt.Println("Error creating")
			return
		}
		status = "open"
	} else {
		status = "open"
	}

	// Вставка нового ордера в БД
	var reqBD string = "INSERT INTO order VALUES ('" + userID + "', '" + strconv.Itoa(req.PairID) + "', '" + strconv.FormatFloat(newQuant, 'f', -1, 64) + "', '" + strconv.FormatFloat(req.Price, 'f', -1, 64) + "', '" + req.Type + "', '" + status + "')"
	_, err2 := zapros.RquestDataBase(reqBD)
	if err2 != nil {

		return
	}

	// Получаем order_id (предполагается, что это последний ордер, добавленный в БД)
	reqBD = "SELECT order.order_id FROM order WHERE order.user_id = '" + userID + "' AND order.closed = '" + status + "'"
	orderIDall, err3 := zapros.RquestDataBase(reqBD)
	if err3 != nil {
		return
	}
	orderID := strings.Split(orderIDall, " \n")
	resOrderID, _ := strconv.Atoi(orderID[len(orderID)-2])

	// Формируем и отправляем JSON-ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateOrderResponseStruct{
		OrderID: resOrderID,
	})
}

func GetOrders(w http.ResponseWriter, r *http.Request) {
	reqBD := "SELECT * FROM order WHERE order.closed = 'open'"

	// Имитируем вызов базы данных
	response, err := zapros.RquestDataBase(reqBD)
	if err != nil {
		http.Error(w, "Ошибка запроса к базе данных", http.StatusInternalServerError)
		return
	}

	// Преобразуем ответ базы данных в строки
	rows := strings.Split(strings.TrimSpace(response), "\n") // Разделяем строки

	// Массив для хранения ордеров
	var orders []GetOrderResponseStruct

	// Парсим каждую строку
	for _, row := range rows {
		fields := strings.Split(row, " ")
		if len(fields) < 7 {
			continue // Пропускаем строки с недостаточным количеством полей
		}

		// Преобразуем каждое поле и заполняем структуру
		orderID, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
		userID, _ := strconv.Atoi(strings.TrimSpace(fields[1]))
		pairID, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
		quantity, _ := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
		orderType := strings.TrimSpace(fields[5])
		price, _ := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64)
		closed := strings.TrimSpace(fields[6])

		order := GetOrderResponseStruct{
			OrderID:  orderID,
			UserID:   userID,
			PairID:   pairID,
			Quantity: quantity,
			Type:     orderType,
			Price:    price,
			Closed:   closed,
		}

		orders = append(orders, order) // Добавляем ордер в массив
	}

	// Устанавливаем заголовки ответа
	w.Header().Set("Content-Type", "application/json")

	// Кодируем массив ордеров в JSON и отправляем клиенту
	json.NewEncoder(w).Encode(orders)
}

func GetAllOrders(w http.ResponseWriter, r *http.Request) {
	reqBD := "SELECT * FROM order"

	// Имитируем вызов базы данных
	response, err := zapros.RquestDataBase(reqBD)
	if err != nil {
		http.Error(w, "Ошибка запроса к базе данных", http.StatusInternalServerError)
		fmt.Println("Error creating request for db: ", err)
		return
	}

	// Преобразуем ответ базы данных в строки
	rows := strings.Split(strings.TrimSpace(response), "\n") // Разделяем строки

	// Массив для хранения ордеров
	var orders []GetOrderResponseStruct

	// Парсим каждую строку
	for _, row := range rows {
		fields := strings.Split(row, " ")
		if len(fields) < 7 {
			continue // Пропускаем строки с недостаточным количеством полей
		}

		// Преобразуем каждое поле и заполняем структуру
		orderID, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
		userID, _ := strconv.Atoi(strings.TrimSpace(fields[1]))
		pairID, _ := strconv.Atoi(strings.TrimSpace(fields[2]))
		quantity, _ := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
		orderType := strings.TrimSpace(fields[5])
		price, _ := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64)
		closed := strings.TrimSpace(fields[6])

		order := GetOrderResponseStruct{
			OrderID:  orderID,
			UserID:   userID,
			PairID:   pairID,
			Quantity: quantity,
			Type:     orderType,
			Price:    price,
			Closed:   closed,
		}

		orders = append(orders, order) // Добавляем ордер в массив
	}

	// Устанавливаем заголовки ответа
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("")
	// Кодируем массив ордеров в JSON и отправляем клиенту
	json.NewEncoder(w).Encode(orders)
}

func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	// Получаем ключ пользователя из заголовка запроса
	userKey := r.Header.Get("X-USER-KEY")
	if userKey == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Проверка наличия ключа пользователя в БД
	var reqUserID string = "SELECT user.user_id FROM user WHERE user.key = '" + userKey + "'"
	userID, err := zapros.RquestDataBase(reqUserID)
	if err != nil || userID == "" {
		http.Error(w, "User unauthorized", http.StatusUnauthorized)
		return
	}

	// Парсинг запроса
	var req DeleteOrderStruct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Проверка является ли пользователь создателем запроса
	userID = userID[:len(userID)-2] // Убираем лишние символы из строки

	var reqUserOrder string = "SELECT * FROM order WHERE order.order_id = '" + strconv.Itoa(req.OrderID) + "' AND order.user_id = '" + userID + "'" //"AND order.closed = 'open'"
	check, err2 := zapros.RquestDataBase(reqUserOrder)
	if err2 != nil || check == "" {
		http.Error(w, "Unautorized access", http.StatusUnauthorized)
		return
	}

	var reqCheckClose string = "SELECT * FROM order WHERE order.closed = 'open' AND order.order_id = '" + strconv.Itoa(req.OrderID) + "'"
	checkClose, err5 := zapros.RquestDataBase(reqCheckClose)
	if err5 != nil || checkClose == "" {
		http.Error(w, "Order is closed", http.StatusUnauthorized)
		return
	}

	// Разбиваем результат запроса на поля
	balanceFields := strings.Split(check, " ")
	if len(balanceFields) < 7 {
		return
	}

	// Удаляем ордер из БД
	var reqBD string = "DELETE FROM order WHERE order.order_id = '" + strconv.Itoa(req.OrderID) + "'"
	_, err3 := zapros.RquestDataBase(reqBD)
	if err3 != nil {
		return
	}

	// Возвращаем деньги обратно на счет пользователю
	floatQuant, _ := strconv.ParseFloat(strings.TrimSpace(balanceFields[3]), 64)
	floatPrice, _ := strconv.ParseFloat(strings.TrimSpace(balanceFields[4]), 64)
	num, _ := strconv.Atoi(balanceFields[2])
	payErr := orderlogic.PayByOrder(userID, num, floatQuant, floatPrice, balanceFields[5], false)
	if payErr != nil {
		http.Error(w, "Not enough funds", http.StatusPaymentRequired)
		return
	}

	// Формируем и отправляем JSON-ответ клиенту
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeleteOrderStruct{
		OrderID: req.OrderID,
	})
}

// пары
type PairResponse struct {
	Pair_id     int `json:"pair_id"`
	Sale_lot_id int `json:"sale_lot_id"`
	Buy_lot_id  int `json:"buy_lot_id"`
}

// Получение информации о парах
func HandlePair(w http.ResponseWriter, r *http.Request) {
	// SQL-запрос для получения всех пар из базы данных
	var getPairsQuery string = "SELECT * FROM pair"

	// Выполняем запрос к базе данных
	dbResponse, err := zapros.RquestDataBase(getPairsQuery)
	if err != nil {
		fmt.Println("Ошибка при получении пар: ", err)
		return // Если произошла ошибка, завершаем выполнение
	}

	// Разделяем ответ базы данных на строки
	dbRows := strings.Split(strings.TrimSpace(dbResponse), "\n") // Каждая строка соответствует записи

	// Массив для хранения данных о парах
	var pairs []PairResponse

	// Обрабатываем каждую строку ответа
	for _, dbRow := range dbRows {
		// Разделяем строку на отдельные поля
		fields := strings.Split(dbRow, " ")
		if len(fields) < 3 {
			continue // Пропускаем строки, где недостаточно данных
		}

		// Преобразуем данные из строки в нужный формат
		pairID, _ := strconv.Atoi(strings.TrimSpace(fields[0]))    // Идентификатор пары
		saleLotID, _ := strconv.Atoi(strings.TrimSpace(fields[1])) // Идентификатор продаваемого лота
		buyLotID, _ := strconv.Atoi(strings.TrimSpace(fields[2]))  // Идентификатор покупаемого лота

		// Создаем структуру для пары
		pairResponse := PairResponse{
			Pair_id:     pairID,
			Sale_lot_id: saleLotID,
			Buy_lot_id:  buyLotID,
		}

		// Добавляем структуру в массив ответов
		pairs = append(pairs, pairResponse)
	}

	// Устанавливаем заголовок ответа и отправляем данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Printf("Пары успешно получены: %+v\n", pairs)
	json.NewEncoder(w).Encode(pairs)
}

//пользователь

// Структура для запроса создания пользователя
type CreateUserRequest struct {
	Username string `json:"username"`
}

// Структура ответа при создании пользователя
type CreateUserResponse struct {
	Key string `json:"key"`
}

// Функция для создания пользователя
func HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Парсинг JSON-запроса от клиента
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utilities.SendJSONError(w, "Ошибка при разборе JSON", http.StatusBadRequest)
		fmt.Println("Ошибка при разборе JSON: ", err)
		return
	}

	// Проверка наличия пользователя в таблице
	checkUserQuery := "SELECT user.username FROM user WHERE user.username = '" + req.Username + "'"
	userCheck, err := zapros.RquestDataBase(checkUserQuery)
	if err != nil {
		utilities.SendJSONError(w, "Ошибка при проверке пользователя", http.StatusInternalServerError)
		fmt.Println("Ошибка при проверке пользователя: ", err)
		return
	}

	// Если есть хотя бы одна строка, значит пользователь уже существует
	if userCheck != "" {
		utilities.SendJSONError(w, "Username занят другим пользователем", http.StatusConflict)
		fmt.Println("Username " + req.Username + " already exists")
		return
	}

	userKey := uuid.New().String()

	var reqBD string = "INSERT INTO user VALUES ('" + req.Username + "', '" + userKey + "')"

	_, err = zapros.RquestDataBase(reqBD)
	if err != nil {
		utilities.SendJSONError(w, "Ошибка при создании пользователя", http.StatusInternalServerError)
		fmt.Println("Ошибка при создании пользователя: ", err)
		return
	}

	// Генерация активов пользователя
	utilities.GenerateMoney(userKey)

	fmt.Println("Пользователь " + req.Username + " создан успешно")

	// Формируем и отправляем JSON-ответ клиенту
	utilities.SendJSONResponse(w, CreateUserResponse{Key: userKey}, http.StatusCreated)
}
