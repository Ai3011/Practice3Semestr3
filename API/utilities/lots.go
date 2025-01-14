package utilities

import (
	"API/zapros"
	"fmt"
)

// Добавление лотов в таблицу lot и pair
func InitLots(lotNames []string) {
	// Проверяем и добавляем отсутствующие лоты в базу данных
	for _, lotName := range lotNames {
		// SQL-запрос для проверки наличия лота в базе данных
		var checkLotQuery string = "SELECT * FROM lot WHERE lot.name = '" + lotName + "'"
		dbResponse, err := zapros.RquestDataBase(checkLotQuery)
		if err != nil {
			fmt.Println("Error checking lot:", err)
			return // Если ошибка, выходим из функции
		} else if dbResponse == "" {
			// SQL-запрос для вставки нового лота, если его нет
			var insertLotQuery string = "INSERT INTO lot VALUES ('" + lotName + "')"
			_, insertErr := zapros.RquestDataBase(insertLotQuery)
			if insertErr != nil {
				fmt.Println("Error inserting lot:", insertErr)
				return // Если ошибка при добавлении, выходим из функции
			}
		}
	}

	// Массив для хранения идентификаторов всех лотов
	var lotIDs []string

	// Получаем идентификаторы лотов для всех имен
	for _, lotName := range lotNames {
		// SQL-запрос для получения идентификатора лота
		var getLotIDQuery string = "SELECT lot.lot_id FROM lot WHERE lot.name = '" + lotName + "'"
		dbResponse, err := zapros.RquestDataBase(getLotIDQuery)
		if err != nil {
			fmt.Println("Error getting lot ID:", err)
			return // Если ошибка, выходим из функции
		}

		// Удаляем лишние символы из ответа базы данных
		dbResponse = dbResponse[:len(dbResponse)-2]
		lotIDs = append(lotIDs, dbResponse) // Добавляем идентификатор в массив
	}

	// Проверяем существование пар и добавляем отсутствующие пары
	for i := 0; i < len(lotIDs); i++ {
		for j := 0; j < len(lotIDs); j++ {
			if i != j { // Пропускаем пары, где лоты совпадают
				// SQL-запрос для проверки существования пары в базе данных
				var checkPairQuery string = "SELECT * FROM pair WHERE pair.first_lot_id = '" + lotIDs[i] + "' AND pair.second_lot_id = '" + lotIDs[j] + "'"
				dbResponse, err := zapros.RquestDataBase(checkPairQuery)
				if err != nil {
					fmt.Println("Error checking pair:", err)
					return // Если ошибка, выходим из функции
				} else if dbResponse == "" {
					// SQL-запрос для вставки новой пары, если она отсутствует
					var insertPairQuery string = "INSERT INTO pair VALUES ('" + lotIDs[i] + "', '" + lotIDs[j] + "')"
					_, insertErr := zapros.RquestDataBase(insertPairQuery)
					if insertErr != nil {
						fmt.Println("Error inserting pair:", insertErr)
						return // Если ошибка при добавлении, выходим из функции
					}
				}
			}
		}
	}
}
