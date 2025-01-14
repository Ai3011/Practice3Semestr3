package main

import (
	"API/config"
	"API/operations"
	"API/utilities"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	// формирование таблицы с парами
	pairList, _, port, _ := config.ConfigRead()
	utilities.InitLots(pairList)

	r := mux.NewRouter()

	// Регистрируем обработчики
	r.HandleFunc("/user", operations.HandleCreateUser).Methods("POST")
	r.HandleFunc("/lot", operations.HandleGetLot).Methods("GET")
	r.HandleFunc("/pair", operations.HandlePair).Methods("GET")
	r.HandleFunc("/balance", operations.HandleGetBalance).Methods("GET")

	r.HandleFunc("/order", operations.CreateOrder).Methods("POST")
	r.HandleFunc("/order", operations.GetOrders).Methods("GET")
	r.HandleFunc("/allorder", operations.GetAllOrders).Methods("GET")
	r.HandleFunc("/order", operations.DeleteOrder).Methods("DELETE")

	// Запускаем сервер на порту 8080
	http.ListenAndServe(":8080", r)
	log.Println("Сервер запущен на порту " + strconv.Itoa(port) + " ...")
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))

}
