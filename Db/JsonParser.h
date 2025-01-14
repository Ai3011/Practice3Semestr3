#ifndef JSONPARSER_H
#define JSONPARSER_H

#include <iostream>
#include <string>
#include <filesystem>
#include <fstream>
#include <sstream>
#include <stdexcept>
#include <mutex>
#include "structures.h"
#include "json.hpp"


using json = nlohmann::json;
using namespace std;


// создание директории
void createDirectory(const string& pathToDir) {
    filesystem::path path(pathToDir);
    if (!filesystem::exists(path)) {
        filesystem::create_directories(path);
    }
}

// создание файла с данными
void createFileData(const string& pathToFile, const string& fileName, const string& data, bool isDirectory) {
    filesystem::path path(pathToFile);
    if (filesystem::exists(path / fileName)) {
        if (isDirectory) {
            ifstream file(path / fileName);
            string line;
            getline(file, line);
            if (line == data) { // Данные уже есть в файле
                file.close();
                return;
            }
            file.close();
        } else {
            return;
        }
    }
    // Если данные в файле не совпадают с JSON или отсутствуют
    ofstream lockFile(path / fileName);
    if (lockFile.is_open()) {
        lockFile << data;  // Записываем данные в файл
        lockFile.close();
    } else {
        throw runtime_error("Не удалось создать файл блокировки в директории");
    }
}


// чтение json файла и создание директорий
void readJsonFile(const string& fileName, SchemaInfo& schemaData) {
    ifstream file(schemaData.filepath + "/" + fileName);
    if (!file.is_open()) {
        throw runtime_error("Не удалось открыть " + fileName);
    }

    // чтение json
    json schema;
    file >> schema;

    // чтение имени таблицы
    schemaData.name = schema["name"];
    createDirectory(schemaData.name);

    // чтение максимального количества ключей
    schemaData.tuplesLimit = schema["tuples_limit"];

    // чтение структуры таблицы
    json tableStructure = schema["structure"];
    for (auto& [key, value] : tableStructure.items()) {
        // создание директорий
        createDirectory(schemaData.name + "/" + key);
        MyVector<string>* tempValue = CreateVector<string>(10, 50);
        string colNames = key + "_id";
        AddVector(*tempValue, colNames);  // Для чтения индекса
        for (auto columns : value) {
            colNames += ",";
            string temp = columns;
            colNames += temp;
            AddVector(*tempValue, temp);  // Добавляем имя столбца в вектор
        }
        createFileData(schemaData.name + "/" + key, "1.csv", colNames, true);
        createFileData(schemaData.name + "/" + key, key + "_pk_sequence.txt", "0", false);
        AddMap<string, MyVector<string>*>(*schemaData.jsonStructure, key, tempValue);
        schemaData.tableMutexes[key];
    }

    file.close();
}

#endif // JSONPARSER_H