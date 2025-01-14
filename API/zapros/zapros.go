package zapros

import (
	"API/config"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
)

// Запрос к базе данных
func RquestDataBase(req string) (string, error) {
	// Устанавливаем TCP-соединение с базой данных на порту
	_, dbIP, _, dbPort := config.ConfigRead()

	conn, err := net.Dial("tcp", dbIP+":"+strconv.Itoa(dbPort))
	if err != nil {
		fmt.Println("Не удалось подключиться к базе данных", http.StatusInternalServerError)
		return "", errors.New("не удалось подключиться к базе данных1")
	}
	defer conn.Close() // Закрываем соединение по завершении

	// Отправляем запрос в базу данных
	fmt.Fprintf(conn, "%s", req+"\n") // Добавляем перевод строки, если база ожидает его

	// Читаем ответ от базы данных
	response, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа от базы данных", http.StatusInternalServerError)
		return "", errors.New("не удалось подключиться к базе данных2")
	}
	str := string(response)
	return str, nil
}
