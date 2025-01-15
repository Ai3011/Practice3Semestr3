package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const apiURL = "http://localhost"

// Структуры данных
type Pair struct {
	PairID    int `json:"pair_id"`
	SaleLotID int `json:"sale_lot_id"`
	BuyLotID  int `json:"buy_lot_id"`
}

type Lot struct {
	LotID int    `json:"lot_id"`
	Name  string `json:"name"`
}

type Order struct {
	OrderID  int     `json:"order_id"`
	UserID   int     `json:"user_id"`
	PairID   int     `json:"lot_id"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Type     string  `json:"type"`
	Closed   string  `json:"closed"`
}

type Balance struct {
	LotID    int     `json:"lot_id"`
	Quantity float64 `json:"quantity"`
}

type UserResponse struct {
	Key string `json:"key"`
}

type OrderRequest struct {
	PairID   int     `json:"pair_id"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Type     string  `json:"type"`
}

// Функция для POST-запросов
func postRequest(endpoint string, payload any, apiKey string) ([]byte, error) {
	data, _ := json.Marshal(payload)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL+endpoint, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-USER-KEY", apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Функция для GET-запросов
func getRequest(endpoint string, apiKey string) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiURL+endpoint, nil)
	if apiKey != "" {
		req.Header.Set("X-USER-KEY", apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func main() {
	// Создаем пользователя. Отправляем POST-запрос на создание нового пользователя с именем "user11".
	user := struct {
		Username string `json:"username"`
	}{"pribil_bot2"}
	resp, _ := postRequest("/user", user, "")
	var userResponse UserResponse
	// Получаем API ключ пользователя из ответа.
	json.Unmarshal(resp, &userResponse)
	apiKey := userResponse.Key

	// Получаем доступные торговые пары через GET-запрос.
	pairsResp, _ := getRequest("/pair", apiKey)
	var pairs []Pair
	json.Unmarshal(pairsResp, &pairs)

	// Получаем список лотов через GET-запрос.
	lotsResp, _ := getRequest("/lot", apiKey)
	var lots []Lot
	json.Unmarshal(lotsResp, &lots)

	// Находим ID лота "RUB", который будет использоваться для торгов.
	var rubLotID int
	for _, lot := range lots {
		if lot.Name == "RUB" {
			rubLotID = lot.LotID
			break
		}
	}

	// Фильтруем только те торговые пары, которые используют RUB.
	var rubPairs []Pair
	for _, pair := range pairs {
		if pair.SaleLotID == rubLotID || pair.BuyLotID == rubLotID {
			rubPairs = append(rubPairs, pair)
		}
	}

	// Бесконечный цикл для мониторинга и принятия решений на основе ордеров
	for {
		// Получаем текущий баланс пользователя через GET-запрос.
		balanceResp, _ := getRequest("/balance", apiKey)
		var balances []Balance
		json.Unmarshal(balanceResp, &balances)

		// Строим карту баланса по LotID.
		balanceMap := make(map[int]float64)
		for _, balance := range balances {
			balanceMap[balance.LotID] = balance.Quantity
		}
		// Выводим баланс для лота с ID 1 (RUB).
		fmt.Println("баланс:", balanceMap[1])

		// Получаем список ордеров через GET-запрос.
		ordersResp, _ := getRequest("/order", apiKey)
		var orders []Order
		json.Unmarshal(ordersResp, &orders)

		// Инициализируем переменные для поиска минимальной цены продажи и максимальной цены покупки.
		var minSell float64 = 100000000
		var pairIDSell int
		var quantitySell float64
		var maxBuy float64 = -100000000
		var pairIDBuy int
		var quantityBuy float64
		var averagePrice float64
		var check int = 0

		// Проходим по всем ордерам, чтобы найти наименьшую цену на продажу и наибольшую цену на покупку.
		for _, order := range orders {
			// Проверяем, относится ли ордер к парам, в которых участвует RUB.
			for _, pair := range rubPairs {
				if order.PairID == pair.PairID {
					if order.Type == "sell" {
						// Если ордер на продажу, ищем минимальную цену продажи.
						if order.Price < minSell {
							minSell = order.Price
							pairIDSell = order.PairID
							quantitySell = order.Quantity
						}
					} else if order.Type == "buy" {
						// Если ордер на покупку, ищем максимальную цену покупки.
						if order.Price > maxBuy {
							maxBuy = order.Price
							pairIDBuy = order.PairID
							quantityBuy = order.Quantity
						}
					}
					// Считаем среднюю цену.
					averagePrice += order.Price
					check++
					break
				}
			}
		}

		// Если были найдены ордера, начинаем принимать решение.
		if check != 0 {
			// Вычисляем среднюю цену.
			averagePrice = averagePrice / float64(check)

			// Определяем, какой ордер выгоднее разместить: на покупку или на продажу.
			var order OrderRequest
			if math.Abs(averagePrice-minSell) > math.Abs(averagePrice-maxBuy) && minSell != 100000000 {
				// Если выгоднее купить, выставляем ордер на покупку.
				order = OrderRequest{
					PairID:   pairIDSell,
					Quantity: quantitySell,
					Price:    minSell,
					Type:     "buy",
				}
			} else {
				// Если выгоднее продать, выставляем ордер на продажу.
				order = OrderRequest{
					PairID:   pairIDBuy,
					Quantity: quantityBuy,
					Price:    maxBuy,
					Type:     "sell",
				}
			}

			// Проверка баланса перед размещением ордера.
			pair := getPairByID(pairs, order.PairID)
			if pair != nil {
				saleLotBalance := balanceMap[pair.SaleLotID]
				buyLotBalance := balanceMap[pair.BuyLotID]

				// Если ордер на покупку, проверяем, достаточно ли средств для покупки.
				if order.Type == "buy" && saleLotBalance >= order.Price*order.Quantity {
					_, err := postRequest("/order", order, apiKey)
					if err == nil {
						// Если ордер успешно выставлен, уменьшаем баланс продажи.
						fmt.Printf("Выставлен лот: %v\n", order)
						balanceMap[pair.SaleLotID] -= order.Price * order.Quantity
					}
				} else if order.Type == "sell" && buyLotBalance >= order.Quantity {
					// Если ордер на продажу, проверяем, достаточно ли товара для продажи.
					_, err := postRequest("/order", order, apiKey)
					if err == nil {
						// Если ордер успешно выставлен, уменьшаем баланс покупки.
						fmt.Printf("Выставлен лот: %v\n", order)
						balanceMap[pair.BuyLotID] -= order.Quantity
					}
				} else {
					// Если средств или товара недостаточно, выводим сообщение об ошибке.
					fmt.Println("Недостаточно средств для совершения операции")
				}
			}
		} else {
			// Если ордеров на покупку или продажу нет, выводим сообщение.
			fmt.Println("Нет ордеров на продажу или покупку")
		}

		// Задержка между итерациями цикла (5 секунд).
		time.Sleep(5 * time.Second)
	}
}

// Функция для получения пары по ID.
func getPairByID(pairs []Pair, pairID int) *Pair {
	for _, pair := range pairs {
		if pair.PairID == pairID {
			return &pair
		}
	}
	return nil
}
