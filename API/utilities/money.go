package utilities

import (
	"API/config"
	"API/zapros"
	"fmt"
)

// Генерация активов пользователей
func GenerateMoney(userKey string) {
	// SQL-запрос для получения идентификатора пользователя по его ключу
	var getUserIDQuery string = "SELECT user.user_id FROM user WHERE user.key = '" + userKey + "'"

	// Выполняем запрос к базе данных для получения идентификатора пользователя
	userIDResponse, err := zapros.RquestDataBase(getUserIDQuery)
	if err != nil {
		return
	}

	// Убираем лишние символы из ответа базы данных
	userID := userIDResponse[:len(userIDResponse)-2]

	// Считываем конфигурацию (список лотов, IP базы данных, порты)
	lotNames, _, _, _ := config.ConfigRead()

	// Обрабатываем каждый лот из конфигурации
	for _, lotName := range lotNames {
		// SQL-запрос для получения идентификатора лота по его названию
		var getLotIDQuery string = "SELECT lot.lot_id FROM lot WHERE lot.name = '" + lotName + "'"

		// Выполняем запрос к базе данных для получения идентификатора лота
		lotIDResponse, err := zapros.RquestDataBase(getLotIDQuery)
		if err != nil {
			fmt.Println("Error getting")
			return
		}

		// Убираем лишние символы из ответа базы данных
		lotID := lotIDResponse[:len(lotIDResponse)-2]

		// SQL-запрос для вставки данных в таблицу `user_lot`
		var insertUserLotQuery string = "INSERT INTO user_lot VALUES ('" + userID + "', '" + lotID + "', '1000')"

		// Выполняем запрос к базе данных для добавления записи
		_, insertErr := zapros.RquestDataBase(insertUserLotQuery)
		if insertErr != nil {
			fmt.Println("Error inserting")
			return // Завершаем выполнение при ошибке
		}
	}
}
