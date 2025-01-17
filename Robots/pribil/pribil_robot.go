package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	// Создаем пользователя и получаем API ключ.
	user := struct {
		Username string `json:"username"`
	}{"pribil_bot00"}
	resp, _ := postRequest("/user", user, "")
	var userResponse UserResponse
	json.Unmarshal(resp, &userResponse)
	apiKey := userResponse.Key

	// Получаем доступные торговые пары и список лотов.
	pairsResp, _ := getRequest("/pair", apiKey)
	var pairs []Pair
	json.Unmarshal(pairsResp, &pairs)

	lotsResp, _ := getRequest("/lot", apiKey)
	var lots []Lot
	json.Unmarshal(lotsResp, &lots)

	// Ищем лот RUB.
	var rubLotID int
	for _, lot := range lots {
		if lot.Name == "RUB" {
			rubLotID = lot.LotID
			break
		}
	}

	// Фильтруем пары с RUB.
	var rubPairs []Pair
	for _, pair := range pairs {
		if pair.SaleLotID == rubLotID || pair.BuyLotID == rubLotID {
			rubPairs = append(rubPairs, pair)
		}
	}

	// Бесконечный цикл для мониторинга.
	for {
		// Получаем баланс пользователя.
		balanceResp, _ := getRequest("/balance", apiKey)
		var balances []Balance
		json.Unmarshal(balanceResp, &balances)

		balanceMap := make(map[int]float64)
		for _, balance := range balances {
			balanceMap[balance.LotID] = balance.Quantity
		}

		// Получаем список ордеров.
		ordersResp, _ := getRequest("/order", apiKey)
		var orders []Order
		json.Unmarshal(ordersResp, &orders)

		// Ищем минимальную цену продажи и максимальную цену покупки.
		var minSellOrder, maxBuyOrder *Order
		for _, order := range orders {
			for _, pair := range rubPairs {
				if order.PairID == pair.PairID {
					if order.Type == "sell" && (minSellOrder == nil || order.Price < minSellOrder.Price) {
						minSellOrder = &order
					} else if order.Type == "buy" && (maxBuyOrder == nil || order.Price > maxBuyOrder.Price) {
						maxBuyOrder = &order
					}
				}
			}
		}

		// Размещаем ордер на покупку или продажу.
		if minSellOrder != nil && maxBuyOrder != nil {
			// Если есть и покупка, и продажа, выбираем более выгодный ордер.
			if minSellOrder.Price < maxBuyOrder.Price {
				placeOrder(minSellOrder, "buy", balanceMap, pairs, apiKey)
			} else {
				placeOrder(maxBuyOrder, "sell", balanceMap, pairs, apiKey)
			}
		} else if minSellOrder != nil {
			placeOrder(minSellOrder, "buy", balanceMap, pairs, apiKey)
		} else if maxBuyOrder != nil {
			placeOrder(maxBuyOrder, "sell", balanceMap, pairs, apiKey)
		} else {
			fmt.Println("Нет подходящих ордеров для торговли.")
		}

		// Задержка между итерациями.
		time.Sleep(5 * time.Second)
	}
}

// Функция для размещения ордера.
func placeOrder(order *Order, orderType string, balanceMap map[int]float64, pairs []Pair, apiKey string) {
	pair := getPairByID(pairs, order.PairID)
	if pair == nil {
		return
	}

	if orderType == "buy" {
		if balanceMap[pair.SaleLotID] >= order.Price*order.Quantity {
			newOrder := OrderRequest{
				PairID:   order.PairID,
				Quantity: order.Quantity,
				Price:    order.Price,
				Type:     "buy",
			}
			postRequest("/order", newOrder, apiKey)
			fmt.Printf("Выставлен ордер на покупку: %+v\n", newOrder)
		} else {
			fmt.Println("Недостаточно средств для покупки.")
		}
	} else if orderType == "sell" {
		if balanceMap[pair.BuyLotID] >= order.Quantity {
			newOrder := OrderRequest{
				PairID:   order.PairID,
				Quantity: order.Quantity,
				Price:    order.Price,
				Type:     "sell",
			}
			postRequest("/order", newOrder, apiKey)
			fmt.Printf("Выставлен ордер на продажу: %+v\n", newOrder)
		} else {
			fmt.Println("Недостаточно средств для продажи.")
		}
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
