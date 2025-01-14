package config

import (
	"encoding/json"
	"log"
	"os"
)

// Структура для хранения данных из JSON
type ConfigStruct struct {
	Lots         []string `json:"lots"`
	DatabaseIP   string   `json:"database_ip"`
	APIPort      int      `json:"api_port"`
	DatabasePort int      `json:"database_port"`
}

// Читаем данные из конфигурационного файла и возвращаем их в виде массива лотов, IP базы данных, порт
func ConfigRead() ([]string, string, int, int) {
	// Читаем содержимое файла конфигурации
	fileContent, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Не удалось открыть файл конфигурации: %v", err) // Завершаем выполнение при ошибке
	}

	// Структура для хранения данных из конфигурационного файла
	var configData ConfigStruct

	// Парсим JSON из файла в структуру
	if err := json.Unmarshal(fileContent, &configData); err != nil {
		log.Fatalf("Ошибка при парсинге JSON: %v", err) // Завершаем выполнение при ошибке
	}

	// Возвращаем значения из структуры конфигурации
	return configData.Lots, configData.DatabaseIP, configData.APIPort, configData.DatabasePort
}
