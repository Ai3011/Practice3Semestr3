#include <fstream>
#include <sstream>
#include <iostream>
#include <stdexcept>
#include <mutex>
#include <string>
#include <filesystem>

#include "structures.h"
#include "JsonParser.h"
#include "Utilities.h"
#include <sys/socket.h>


using namespace std;


// удаление опострафа и проверка синтаксиса
string cleanText(string& str) {
    if (str[str.size() - 1] == ',' && str[str.size() - 2] == ')') {
        str = getSubsting(str, 0, str.size() - 2);
    } else if (str[str.size() - 1] == ',' || str[str.size() - 1] == ')') {
        str = getSubsting(str, 0, str.size() - 1);
    }

    if (str[0] == '\'' && str[str.size() - 1] == '\'') {
        str = getSubsting(str, 1, str.size() - 1);
        return str;
    } else {
        throw runtime_error("invalid sintaxis in VALUES " + str);
    }
}

// проверка количества аргументов относительно столбцов таблиц
void Validate(int colLen, const MyVector<string>& namesOfTable, const MyHashMap<string, MyVector<string>*>& jsonStructure) {
    for (int i = 0; i < namesOfTable.length; i++) {
        MyVector<string>* temp = GetMap<string, MyVector<string>*>(jsonStructure, namesOfTable.data[i]);
        if (temp->length - 1 != colLen) {      // добавить удаление первого элемента из мапа
            throw runtime_error("the number of arguments is not equal to the columns in " + namesOfTable.data[i]);
        }
    }
}

// чтение файла с количеством записей и перезапись
int readPrKey(const string& path, const bool record, const int newID) {
    fstream pkFile(path);
    if (!pkFile.is_open()) {
        throw runtime_error("Не удалось открыть" + path);
    }
    int lastID = 0;
    if (record) {
        pkFile << newID;
    } else {
        pkFile >> lastID;
    }
    pkFile.close();
    return lastID;
}

// добавление строк в файл
void insertRows(MyVector<MyVector<string>*>& addNewData, MyVector<string>& namesOfTable, SchemaInfo& dataOfSchema) {
    for (int i = 0; i < namesOfTable.length; i++) {
        string pathToCSV = dataOfSchema.filepath + "/" + dataOfSchema.name + "/" + namesOfTable.data[i];
        int lastID = 0;

        // Захватываем мьютекс для таблицы, если она существует в tableMutexes
        auto mutexIt = dataOfSchema.tableMutexes.find(namesOfTable.data[i]);
        if (mutexIt != dataOfSchema.tableMutexes.end()) {
            unique_lock<mutex> lock(mutexIt->second); // Блокировка мьютекса
            cout << "mutex is locked " << namesOfTable.data[i] << endl;

            try {
                lastID = readPrKey(pathToCSV + "/" + namesOfTable.data[i] + "_pk_sequence.txt", false, 0);
            } catch (const exception& err) {
                throw runtime_error(err.what());
                return;
            }

            int newID = lastID;
            for (int j = 0; j < addNewData.length; j++) {
                newID++;
                string tempPath;
                if (lastID / dataOfSchema.tuplesLimit < newID / dataOfSchema.tuplesLimit) {
                    tempPath = pathToCSV + "/" + to_string(newID / dataOfSchema.tuplesLimit + 1) + ".csv";
                } else {
                    tempPath = pathToCSV + "/" + to_string(lastID / dataOfSchema.tuplesLimit + 1) + ".csv";
                }
                fstream csvFile(tempPath, ios::app);
                if (!csvFile.is_open()) {
                    throw runtime_error("Failed to open" + tempPath);
                }
                csvFile << endl << newID;
                for (int k = 0; k < addNewData.data[j]->length; k++) {
                    csvFile << "," << addNewData.data[j]->data[k];
                }
                csvFile.close();
            }
            readPrKey(pathToCSV + "/" + namesOfTable.data[i] + "_pk_sequence.txt", true, newID);
        }
    }
}

// разделение запроса вставки на части
void parseInsert(const MyVector<string>& slovs, SchemaInfo& dataOfSchema) {
    MyVector<string>* namesOfTables = CreateVector<string>(5, 50);
    MyVector<MyVector<string>*>* addNewData = CreateVector<MyVector<string>*>(10, 50);
    bool afterValues = false;
    int countTabNames = 0;
    int countAddData = 0;
    for (int i = 2; i < slovs.length; i++) {
        if (slovs.data[i][slovs.data[i].size() - 1] == ',') {
            slovs.data[i] = getSubsting(slovs.data[i], 0, slovs.data[i].size() - 1);
        }
        if (slovs.data[i] == "VALUES") {
            afterValues = true;
        } else if (afterValues) {
            countAddData++;
            if (slovs.data[i][0] == '(') {
                MyVector<string>* tempData = CreateVector<string>(5, 50);
                slovs.data[i] = getSubsting(slovs.data[i], 1, slovs.data[i].size());

                while (slovs.data[i][slovs.data[i].size() - 1] != ')' && slovs.data[i][slovs.data[i].size() - 2] != ')') {
                    try {
                        cleanText(slovs.data[i]);
                    } catch (const exception& err) {
                        throw runtime_error(err.what());
                        return;
                    }
                    
                    AddVector<string>(*tempData, slovs.data[i]);
                    i++;
                }
                try {
                    cleanText(slovs.data[i]);
                    AddVector<string>(*tempData, slovs.data[i]);
                    Validate(tempData->length, *namesOfTables, *dataOfSchema.jsonStructure);
                } catch (const exception& err) {
                    throw runtime_error(err.what());
                    return;
                }
                AddVector<MyVector<string>*>(*addNewData, tempData);
            }
            
        } else {
            countTabNames++;
            try {
                GetMap(*dataOfSchema.jsonStructure, slovs.data[i]);
            } catch (const exception& err) {
                throw runtime_error(err.what());
                return;
            }
            AddVector<string>(*namesOfTables, slovs.data[i]);
        }
    }
    if (countTabNames == 0 || countAddData == 0) {
        throw runtime_error("missing table name or data in VALUES");
    }

    try {
        insertRows(*addNewData, *namesOfTables, dataOfSchema);
    } catch (const exception& err) {
        throw runtime_error(err.what());
        return;
    }
}

// Вспомогательная функция для разделения строки по оператору
MyVector<MyVector<string>*>* splitByOperator(const MyVector<string>& query, const string& op) {
    MyVector<string>* left = CreateVector<string>(6, 50);
    MyVector<string>* right = CreateVector<string>(6, 50);
    bool afterOp = false;
    for (int i = 0; i < query.length; i++) {
        if (query.data[i] == op) {
            afterOp = true;
        } else if (afterOp) {
            AddVector(*right, query.data[i]);
        } else {
            AddVector(*left, query.data[i]);
        }
    }
    MyVector<MyVector<string>*>* parseVector = CreateVector<MyVector<string>*>(5, 50);
    if (afterOp) {
        AddVector(*parseVector, left);
        AddVector(*parseVector, right);
        
    } else {
        AddVector(*parseVector, left);
    }
    return parseVector;
}


string SanitizeText(string str) {
    if (str[0] == '\'' && str[str.size() - 1] == '\'') {
        str = getSubsting(str, 1, str.size() - 1);
        return str;
    } else {
        throw runtime_error("Неверный синтаксис в WHERE " + str);
    }
}

bool isValidRow(Node* node, const MyVector<string>& row, const MyHashMap<string, MyVector<string>*>& jsonStructure, const string& namesOfTable) {
    if (!node) {
        return false;
    }

    switch (node->nodeType) {
    case NodeType::ConditionNode: {
        if (node->value.length != 3) {
            return false;
        }

        MyVector<string> *part1Splitted = splitRow(node->value.data[0], '.');
        if (part1Splitted->length != 2) {
            return false;
        }
    
        // существует ли запрашиваемая таблица
        int columnIndex = -1;
        try {
            MyVector<string>* colNames = GetMap(jsonStructure, part1Splitted->data[0]);
            for (int i = 0; i < colNames->length; i++) {
                if (colNames->data[i] == part1Splitted->data[1]) {
                    columnIndex = i;
                    break;
                }
            }
        } catch (const exception& e) {
            throw runtime_error(e.what());
            return false;
        }

        if (columnIndex == -1) {
            cerr << "Column " << part1Splitted->data[1] << " is missing in table " << part1Splitted->data[0] << std::endl;
            return false;
        }

        string delApostr = SanitizeText(node->value.data[2]);
        if (namesOfTable == part1Splitted->data[0] && row.data[columnIndex] == delApostr) {  
            return true;
        }

        return false;
    }
    case NodeType::OrNode:
        return isValidRow(node->left, row, jsonStructure, namesOfTable) ||
                isValidRow(node->right, row, jsonStructure, namesOfTable);
    case NodeType::AndNode:
        return isValidRow(node->left, row, jsonStructure, namesOfTable) &&
                isValidRow(node->right, row, jsonStructure, namesOfTable);
    default:
        return false;
    }
}


Node* getConditionTree(const MyVector<string>& query) {
    MyVector<MyVector<string>*>* orParts = splitByOperator(query, "OR");

    // Если найден оператор OR
    if (orParts->length > 1) {
        Node* root = new Node(NodeType::OrNode);
        root->left = getConditionTree(*orParts->data[0]);
        root->right = getConditionTree(*orParts->data[1]);
        return root;
    }
    // Если найден оператор AND
    MyVector<MyVector<std::string>*>* andParts = splitByOperator(query, "AND");
    if (andParts->length > 1) {
        Node* root = new Node(NodeType::AndNode);
        root->left = getConditionTree(*andParts->data[0]);
        root->right = getConditionTree(*andParts->data[1]);
        return root;
    }

    // Если это простое условие
    return new Node(NodeType::ConditionNode, query);
}



// перезапись во временный файл информации кроме удаленной
void removeData(MyVector<string>& namesOfTable, MyVector<string>& listOfCondition, SchemaInfo& dataOfShema) {
    Node* nodeWere = getConditionTree(listOfCondition);
     
    for (int i = 0; i < namesOfTable.length; i++) {
        int fileIndex = 1;
        string pathToCSV = dataOfShema.filepath + "/" + dataOfShema.name + "/" + namesOfTable.data[i];
        auto mutexIt = dataOfShema.tableMutexes.find(namesOfTable.data[i]);
        if (mutexIt != dataOfShema.tableMutexes.end()) {
            unique_lock<mutex> lock(mutexIt->second); // Блокировка мьютекса

            while (filesystem::exists(pathToCSV + "/" + to_string(fileIndex) + ".csv")) {
                ifstream file(pathToCSV + "/" + to_string(fileIndex) + ".csv");
                if (!file.is_open()) {
                    throw runtime_error("Ошибка открытия файла " + (pathToCSV + "/" + to_string(fileIndex) + ".csv"));
                }
                ofstream tempFile(pathToCSV + "/" + to_string(fileIndex) + "_temp.csv");

                string line;
                getline(file, line);
                tempFile << line;
                while (getline(file, line)) {
                    MyVector<string>* row = splitRow(line, ',');
                    try {
                        if (!isValidRow(nodeWere, *row, *dataOfShema.jsonStructure, namesOfTable.data[i])) {
                            tempFile << endl << line;
                        }
                    } catch (const exception& e) {
                        tempFile.close();
                        file.close();
                        remove((pathToCSV + "/" + to_string(fileIndex) + "_temp.csv").c_str());
                        throw runtime_error(e.what());
                        return;
                    }
                }
                tempFile.close();
                file.close();
                if (remove((pathToCSV + "/" + to_string(fileIndex) + ".csv").c_str()) != 0) {
                    throw runtime_error("Error deleting file");
                    return;
                }
                if (rename((pathToCSV + "/" + to_string(fileIndex) + "_temp.csv").c_str(), (pathToCSV + "/" + to_string(fileIndex) + ".csv").c_str()) != 0) {
                    throw runtime_error("Error renaming file");
                    return;
                }

                fileIndex++;
            }
        }
    }
}

// разбиение запроса удаления на кусочки
void parseDelete(const MyVector<string>& words, SchemaInfo& dataOfSchema) {
    MyVector<string>* namesOfTable = CreateVector<string>(5, 50);
    MyVector<string>* listOfCondition = CreateVector<string>(5, 50);
    int countTabNames = 0;
    int countWereData = 0;
    bool afterWhere = false;
    for (int i = 2; i < words.length; i++ ) {
        if (words.data[i][words.data[i].size() - 1] == ',') {
            words.data[i] = getSubsting(words.data[i], 0, words.data[i].size() - 1);
        }
        if (words.data[i] == "WHERE") {
            afterWhere = true;
        } else if (afterWhere) {
            AddVector<string>(*listOfCondition, words.data[i]);
            countWereData++;
        } else {
            countTabNames++;
            try {
                GetMap(*dataOfSchema.jsonStructure, words.data[i]);
            } catch (const exception& e) {
                throw runtime_error(e.what());
                return;
            }
            AddVector<string>(*namesOfTable, words.data[i]);
        }
    }
    if (countTabNames == 0 || countWereData == 0) {
        throw runtime_error("missing table name or data in WERE");
    }

    try {
        removeData(*namesOfTable, *listOfCondition, dataOfSchema);
    } catch (const exception& err) {
        throw;
        return;
    }
}



bool writeAllRows(Node* nodeWere, const string& nameOfTable , string& line, MyVector<MyVector<string>*>& tabData, SchemaInfo& schemaData, bool whereValue) {
    MyVector<string>* row = splitRow(line, ',');
    if (whereValue) {
        try {
            if (isValidRow(nodeWere, *row, *schemaData.jsonStructure, nameOfTable)) {
                AddVector(tabData, row);
            }
        } catch (const exception& err) {
            throw;
            return false;
        }
    } else {
        AddVector(tabData, row);
    }
    return true;
}

// считывание подходящих строк из выбранных столбцов
bool writePhRows(Node* nodeWere, const string& nameOfTable, string& line, MyVector<MyVector<string>*>& tabData, SchemaInfo& dataOfSchema, bool whereValue, MyVector<int>& colIndex) {
    MyVector<string>* row = splitRow(line, ',');
    MyVector<string>* newRow = CreateVector<string>(colIndex.length, 50);
    if (whereValue) {
        try {
            if (isValidRow(nodeWere, *row, *dataOfSchema.jsonStructure, nameOfTable)) {
                for (int i = 0; i < colIndex.length; i++) {
                    AddVector(*newRow, row->data[colIndex.data[i]]);
                }
                AddVector(tabData, newRow);
            }
        } catch (const exception& err) {
            throw;
            return false;
        }
    } else {
        for (int i = 0; i < colIndex.length; i++) {
            AddVector(*newRow, row->data[colIndex.data[i]]);
        }
        AddVector(tabData, newRow);
    }
    return true;
}


// чтение таблицы из файла
MyVector<MyVector<string>*>* ReadTable(const string& nameOfTable, SchemaInfo& dataOfSchema, const MyVector<string>& namesOfColumns, const MyVector<string>& listOfCondition, bool whereValue) {
    MyVector<MyVector<string>*>* tabData = CreateVector<MyVector<string>*>(5, 50);
    string pathToCSV = dataOfSchema.filepath + "/" + dataOfSchema.name + "/" + nameOfTable;
    int fileIndex = 1;

    // Захватываем мьютекс для таблицы, если она существует в tableMutexes
    auto mutexIt = dataOfSchema.tableMutexes.find(nameOfTable);
    if (mutexIt != dataOfSchema.tableMutexes.end()) {
        unique_lock<mutex> lock(mutexIt->second); // Блокировка мьютекса
        
        Node* nodeWere = getConditionTree(listOfCondition);
        while (filesystem::exists(pathToCSV + "/" + to_string(fileIndex) + ".csv")) {
            ifstream file(pathToCSV + "/" + to_string(fileIndex) + ".csv");
            if (!file.is_open()) {
                throw runtime_error("Ошибка открытия файла: " + (pathToCSV + "/" + to_string(fileIndex) + ".csv"));
            }
            string firstLine;
            getline(file, firstLine);
            if (namesOfColumns.data[0] == "*") {
                string line;
                while (getline(file, line)) {
                    if (!writeAllRows(nodeWere, nameOfTable, line, *tabData, dataOfSchema, whereValue)) {
                        file.close();
                        return tabData;
                    }
                }
            } else {
                MyVector<string>* filenamesOfColumns = GetMap<string, MyVector<string>*>(*dataOfSchema.jsonStructure, nameOfTable);
                MyVector<int>* colIndex = CreateVector<int>(10, 50);
                for (int i = 0; i < filenamesOfColumns->length; i++) {
                    for (int j = 1; j < namesOfColumns.length; j++) {
                        if (filenamesOfColumns->data[i] == namesOfColumns.data[j]) {
                            AddVector(*colIndex, i);
                        }
                    }
                }
                string line;
                while (getline(file, line)) {
                    if (!writePhRows(nodeWere, nameOfTable, line, *tabData, dataOfSchema, whereValue, *colIndex)) {
                        file.close();
                        return tabData;
                    }
                }
            }

            file.close();
            fileIndex += 1;
        }
    }
    return tabData;
}


// вывод содержимого таблиц в виде декартового произведения
void cartesianProduct(const MyVector<MyVector<MyVector<string>*>*>& dataOfTables, MyVector<MyVector<string>*>& temp, int counterTab, int tab, int clientSocket) {
    for (int i = 0; i < dataOfTables.data[counterTab]->length; i++) {
        temp.data[counterTab] = dataOfTables.data[counterTab]->data[i];

        if (counterTab < tab - 1) {
            cartesianProduct(dataOfTables, temp, counterTab + 1, tab, clientSocket);
        } else {
            for (int j = 0; j < tab; j++) {
                for (int k = 0; k < temp.data[j]->length; k++) {
                    send(clientSocket, (temp.data[j]->data[k] + " ").c_str(), (temp.data[j]->data[k] + " ").size(), 0);
                }
            }
            string enter = "\n";
            send(clientSocket, enter.c_str(), enter.size(), 0);
        }
    }

    return;
}

// подготовка к чтению и выводу данных
void selectDataPreparation(const MyVector<string>& namesOfColumns, const MyVector<string>& namesOfTables, const MyVector<string>& listOfCondition, SchemaInfo& dataOfSchema, bool whereValue, int clientSocket) {
    MyVector<MyVector<MyVector<string>*>*>* dataOfTables = CreateVector<MyVector<MyVector<string>*>*>(10, 50);
    if (namesOfColumns.data[0] == "*") {      // чтение всех данных из таблиц
        for (int j = 0; j < namesOfTables.length; j++) {
            MyVector<MyVector<string>*>* tableData = ReadTable(namesOfTables.data[j], dataOfSchema, namesOfColumns, listOfCondition, whereValue);
            AddVector(*dataOfTables, tableData);
        }
    } else {
        for (int i = 0; i < namesOfTables.length; i++) {
            MyVector<string>* tabColPair = CreateVector<string>(5, 50);
            AddVector(*tabColPair, namesOfTables.data[i]);
            for (int j = 0; j < namesOfColumns.length; j++) {
                MyVector<string>* splitNamesOfColumns = splitRow(namesOfColumns.data[j], '.');
                try {
                    GetMap(*dataOfSchema.jsonStructure, splitNamesOfColumns->data[0]);
                } catch (const exception& e) {
                    throw runtime_error(e.what());
                    return;
                }
                if (splitNamesOfColumns->data[0] == namesOfTables.data[i]) {
                    AddVector(*tabColPair, splitNamesOfColumns->data[1]);
                }
            }
            MyVector<MyVector<string>*>* tableData = ReadTable(tabColPair->data[0], dataOfSchema, *tabColPair, listOfCondition, whereValue);;
            AddVector(*dataOfTables, tableData);
        }
    }

    MyVector<MyVector<string>*>* temp = CreateVector<MyVector<string>*>(dataOfTables->length * 2, 50);
    string resStr;
    cartesianProduct(*dataOfTables, *temp, 0, dataOfTables->length, clientSocket);
    return;
}

// парсинг SELECT запроса
void parseSelect(const MyVector<string>& slovs, SchemaInfo& dataOfSchema, int clientSocket) {
    MyVector<string>* namesOfColumns = CreateVector<string>(10, 50);          // названия колонок в формате таблица1.колонка1
    MyVector<string>* namesOfTables = CreateVector<string>(10, 50);        // названия таблиц в формате  таблица1
    MyVector<string>* listOfCondition = CreateVector<string>(10, 50);     // список условий where
    bool afterFrom = false;
    bool afterWhere = false;
    int countTabNames = 0;
    int countData = 0;
    int countWhereData = 0;
    for (int i = 1; i < slovs.length; i++) {
        if (slovs.data[i][slovs.data[i].size() - 1] == ',') {
            slovs.data[i] = getSubsting(slovs.data[i], 0, slovs.data[i].size() - 1);
        }
        if (slovs.data[i] == "FROM") {
            afterFrom = true;
        } else if (slovs.data[i] == "WHERE") {
            afterWhere = true;
        } else if (afterWhere) {
            countWhereData++;
            AddVector<string>(*listOfCondition, slovs.data[i]);
        } else if (afterFrom) {
            try {
                GetMap(*dataOfSchema.jsonStructure, slovs.data[i]);
            } catch (const exception& e) {
                throw runtime_error(e.what());
                return;
            }
            countTabNames++;
            AddVector(*namesOfTables, slovs.data[i]);
        } else {
            countData++;
            AddVector(*namesOfColumns, slovs.data[i]);
        }
    }
    if (countTabNames == 0 || countData == 0) {
        throw runtime_error("Отсутствует имя таблицы или данные в FROM");
    }
    if (countWhereData == 0) {
        selectDataPreparation(*namesOfColumns, *namesOfTables, *listOfCondition, dataOfSchema, false, clientSocket);
    } else {
        selectDataPreparation(*namesOfColumns, *namesOfTables, *listOfCondition, dataOfSchema, true, clientSocket);
    }
}


