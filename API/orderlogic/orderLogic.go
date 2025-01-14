package orderlogic

import (
	"API/zapros"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type GetOrderResponse struct {
	OrderID  int     `json:"order_id"`
	UserID   int     `json:"user_id"`
	PairID   int     `json:"lot_id"`
	Quantity float64 `json:"quantity"`
	Type     string  `json:"type"`
	Price    float64 `json:"price"`
	Closed   string  `json:"closed"`
}

// Списание средств
func PayByOrder(userID string, pairID int, payMoney float64, price float64, orderType string, spisanie bool) error {
	// Получить информацию о валютной паре
	reqPair := "SELECT * FROM pair WHERE pair.pair_id = '" + strconv.Itoa(pairID) + "'"
	pairData, err := zapros.RquestDataBase(reqPair)
	if err != nil || pairData == "" {
		return errors.New("валютная пара не найдена")
	}

	// Разбираем данные пары
	pairFields := strings.Split(pairData, " ") // "1 RUB USD"
	if len(pairFields) < 3 {
		return errors.New("некорректные данные пары")
	}
	firstLotID := pairFields[1]
	secondLotID := pairFields[2]

	// Определяем, с какого счета списывать/начислять
	var lotID string
	if orderType == "buy" {
		lotID = secondLotID // Покупка, списываем вторую валюту
	} else if orderType == "sell" {
		lotID = firstLotID // Продажа, списываем первую валюту
	}

	// Получить баланс пользователя
	reqBalance := "SELECT * FROM user_lot WHERE user_lot.user_id = '" + userID + "' AND user_lot.lot_id = '" + lotID + "'"
	balanceData, err := zapros.RquestDataBase(reqBalance)
	if err != nil || balanceData == "" {
		return errors.New("недостаточно средств или запись не найдена")
	}

	balanceFields := strings.Split(balanceData, " ") // "22,2,1,1000"
	if len(balanceFields) < 4 {
		return errors.New("некорректные данные баланса")
	}
	currentBalance, _ := strconv.ParseFloat(balanceFields[3], 64)

	// Проверка средств
	if spisanie && currentBalance < payMoney*price { // добавить домножение
		return errors.New("недостаточно средств")
	}

	// Обновить баланс
	var newBalance float64
	if spisanie {
		newBalance = currentBalance - payMoney*price
	} else {
		newBalance = currentBalance + payMoney*price
	}

	// Удалить старую запись
	reqDelete := "DELETE FROM user_lot WHERE user_lot.user_id = '" + userID + "' AND user_lot.lot_id = '" + lotID + "'"
	_, err = zapros.RquestDataBase(reqDelete)
	if err != nil {
		return errors.New("ошибка при обновлении баланса")
	}

	// Вставить новую запись
	reqInsert := fmt.Sprintf("INSERT INTO user_lot VALUES ('%s', '%s', '%.2f')", userID, lotID, newBalance)
	_, err = zapros.RquestDataBase(reqInsert)
	if err != nil {
		return errors.New("ошибка при обновлении баланса")
	}

	return nil
}

// Проведение транзакций
func conductTransactions(quantity float64, orders []GetOrderResponse, orderType string) (float64, error) {
	totalQuantity := 0.0

	for _, order := range orders {
		if totalQuantity >= quantity { // Если запрос полностью покрыт
			break
		}

		if order.Quantity+totalQuantity > quantity { // Если текущий ордер больше, чем нужно, создаем остаток
			var remainingQuantity float64 = order.Quantity + totalQuantity - quantity
			if remainingQuantity > 0 { // Создаем новый ордер с остатком
				// Удаляем ордер
				var forcloseOrderQuery string = "DELETE FROM order WHERE order.order_id = '" + strconv.Itoa(order.OrderID) + "' AND order.closed = 'open'"
				_, err := zapros.RquestDataBase(forcloseOrderQuery)
				if err != nil {
					return -1, errors.New("ошибка при закрытии ордера")
				}

				// Зачисляем часть денег владельцу ордера (тип ордера противоположный, так как произошло завершение транзакции)
				_ = PayByOrder(strconv.Itoa(order.UserID), order.PairID, remainingQuantity, order.Price, orderType, false)

				// Создаем закрытый ордер с частичной суммой
				var createOrderQueryClose string = "INSERT INTO order VALUES ('" + strconv.Itoa(order.UserID) + "', '" + strconv.Itoa(order.PairID) + "', '" + strconv.FormatFloat(order.Quantity-remainingQuantity, 'f', -1, 64) + "', '" + strconv.FormatFloat(order.Price, 'f', -1, 64) + "', '" + order.Type + "', 'close')"
				_, err = zapros.RquestDataBase(createOrderQueryClose)
				if err != nil {
					return -1, errors.New("ошибка при создании остаточного ордера")
				}

				// Создаем новый ордер с остатком
				var createOrderQuery string = "INSERT INTO order VALUES ('" + strconv.Itoa(order.UserID) + "', '" + strconv.Itoa(order.PairID) + "', '" + strconv.FormatFloat(remainingQuantity, 'f', -1, 64) + "', '" + strconv.FormatFloat(order.Price, 'f', -1, 64) + "', '" + order.Type + "', '" + order.Closed + "')"
				_, err = zapros.RquestDataBase(createOrderQuery)
				if err != nil {
					return -1, errors.New("ошибка при создании остаточного ордера")
				}
			}

			// Уменьшаем количество в текущем ордере до закрытия
			order.Quantity = quantity - totalQuantity
		} else {
			// Удаляем ордер
			var forcloseOrderQuery string = "DELETE FROM order WHERE order.order_id = '" + strconv.Itoa(order.OrderID) + "' AND order.closed = 'open'"
			_, err := zapros.RquestDataBase(forcloseOrderQuery)
			if err != nil {
				return -1, errors.New("ошибка при закрытии ордера")
			}

			// Зачисляем все деньги владельцу ордера (тип ордера противоположный, так как произошло завершение транзакции)
			_ = PayByOrder(strconv.Itoa(order.UserID), order.PairID, order.Quantity, order.Price, orderType, false)

			// Закрываем ордер
			if order.Quantity+totalQuantity <= quantity {
				var closeOrderQuery string = "INSERT INTO order VALUES ('" + strconv.Itoa(order.UserID) + "', '" + strconv.Itoa(order.PairID) + "', '" + strconv.FormatFloat(order.Quantity, 'f', -1, 64) + "', '" + strconv.FormatFloat(order.Price, 'f', -1, 64) + "', '" + order.Type + "', 'close')"
				_, err := zapros.RquestDataBase(closeOrderQuery)
				if err != nil {
					return -1, errors.New("ошибка при закрытии ордера")
				}
			}
		}

		totalQuantity += order.Quantity
	}

	return totalQuantity, nil
}

// Поиск уже существующих ордеров для транзакции
func SearchOrder(searchUserID string, orderPairID int, orderType string, quantity float64, price float64, types string) (float64, error) {
	var searchOrderTypes string
	if types == "buy" {
		searchOrderTypes = "sell"
	} else {
		searchOrderTypes = "buy"
	}

	// Получить все открытые ордера по паре
	reqOrders := "SELECT * FROM order WHERE order.closed = 'open' AND order.pair_id = '" + strconv.Itoa(orderPairID) + "'"
	fmt.Println("запрос ордера", reqOrders)
	orderData, err := zapros.RquestDataBase(reqOrders)
	if err != nil {
		return -1, errors.New("ошибка при поиске ордеров")
	}

	// Разбираем строки
	rows := strings.Split(strings.TrimSpace(orderData), "\n")
	var orders []GetOrderResponse

	for _, row := range rows {
		fields := strings.Split(row, " ")
		if len(fields) < 7 {
			continue
		}

		// Парсим данные
		orderID, _ := strconv.Atoi(fields[0])
		userID, _ := strconv.Atoi(fields[1])
		pairID, _ := strconv.Atoi(fields[2])
		orderQuantity, _ := strconv.ParseFloat(fields[3], 64)
		orderPrice, _ := strconv.ParseFloat(fields[4], 64)
		orderType := fields[5]
		closed := fields[6]

		// Фильтруем ордера по типу и цене
		if (orderType == "buy" && price <= orderPrice || orderType == "sell" && price >= orderPrice) && orderType == searchOrderTypes && strconv.Itoa(userID) != searchUserID {
			orders = append(orders, GetOrderResponse{
				OrderID:  orderID,
				UserID:   userID,
				PairID:   pairID,
				Quantity: orderQuantity,
				Type:     orderType,
				Price:    orderPrice,
				Closed:   closed,
			})
		}
	}

	// Если не найдено подходящих ордеров
	if len(orders) == 0 {
		return quantity, nil
	}

	// Сортировка: для покупки выбираем минимальную цену, для продажи максимальную
	sort.Slice(orders, func(i, j int) bool {
		if orderType == "buy" {
			return orders[i].Price < orders[j].Price
		}
		return orders[i].Price > orders[j].Price
	})

	// Проводим транзакции с подходящими ордерами
	totalQuantity, er := conductTransactions(quantity, orders, orderType)
	if er != nil {
		fmt.Println("Error when searching for orders")
		return -1, errors.New("ошибка при поиске ордеров")
	}

	// Зачисляем все деньги владельцу исходного ордера (тип ордера противоположный, так как произошло завершение транзакции)
	_ = PayByOrder(searchUserID, orderPairID, quantity, price, searchOrderTypes, false)

	// Возвращаем оставшееся количество, если транзакция не полностью выполнена
	if totalQuantity < quantity {
		return quantity - totalQuantity, nil
	}

	return 0, nil
}
