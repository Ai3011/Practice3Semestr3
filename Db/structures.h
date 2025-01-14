#ifndef STRUCTURES_H
#define STRUCTURES_H

#include <iostream>
#include <iomanip>
#include <mutex>
#include <map>
#include <string>

using namespace std;

// структура для хранения значения
template <typename TK, typename TV>
struct NodeMap {
    TK key;
    TV value;
    NodeMap* next;
};

// структура для хранения ключа и значения
template <typename TK, typename TV>
struct MyHashMap {
    NodeMap<TK, TV>** data;
    size_t length;
    size_t capacity;
    size_t loadFactor;
};

// хэш-функция для ключа string
template <typename TK>
int HashCode(const TK& key, const int capacity) {
    unsigned long hash = 5381;
    int c = 0;
    for (char ch : key) {
        hash = ((hash << 5) + hash) + ch;
    }
    return hash % capacity;
}

// инициализация хэш таблицы
template <typename TK, typename TV>
MyHashMap<TK, TV>* CreateMap(int initCapacity, int initLoadFactor) {
    if (initCapacity <= 0 || initLoadFactor <= 0 || initLoadFactor > 100) {
        throw std::runtime_error("Индекс вне диапазона");
    }

    MyHashMap<TK, TV>* map = new MyHashMap<TK, TV>;
    map->data = new NodeMap<TK, TV>*[initCapacity];
    for (size_t i = 0; i < initCapacity; i++) {
        map->data[i] = nullptr;
    }

    map->length = 0;
    map->capacity = initCapacity;
    map->loadFactor = initLoadFactor;
    return map;
}

// расширение
template <typename TK, typename TV>
void Expansion(MyHashMap<TK, TV>& map) {
    size_t newCap = map.capacity * 2;
    NodeMap<TK, TV>** newData = new NodeMap<TK, TV>*[newCap];
    for (size_t i = 0; i < newCap; i++) {
        newData[i] = nullptr;
    }
    // проход по всем ячейкам
    for (size_t i = 0; i < map.capacity; i++) {
        NodeMap<TK, TV>* curr = map.data[i];
        // проход по парам коллизионных значений и обновление
        while (curr != nullptr) {
            NodeMap<TK, TV>* next = curr->next;
            size_t index = HashCode(curr->key, newCap);
            curr->next = newData[index];
            newData[index] = curr;
            curr = next;
        }
    }

    delete[] map.data;

    map.data = newData;
    map.capacity = newCap;
}

// обработка коллизий
template <typename TK, typename TV>
void CollisionManage(MyHashMap<TK, TV>& map, int index, const TK& key, const TV& value) {
    NodeMap<TK, TV>* newNode = new NodeMap<TK, TV>{key, value, nullptr};
    NodeMap<TK, TV>* curr = map.data[index];
    while (curr->next != nullptr) {
        curr = curr->next;
    }
    curr->next = newNode;
}

// добавление элементов
template <typename TK, typename TV>
void AddMap(MyHashMap<TK, TV>& map, const TK& key, const TV& value) {
    if ((map.length + 1) * 100 / map.capacity >= map.loadFactor) {
        Expansion(map);
    }
    size_t index = HashCode(key, map.capacity);
    NodeMap<TK, TV>* temp = map.data[index];
    if (temp != nullptr) {
        while (temp != nullptr) {
            if (temp->key == key) {
                // Элемент уже существует, обновить значение
                temp->value = value;
                map.data[index] = temp;
                return;
            }
            temp = temp->next;
        }
        CollisionManage(map, index, key, value);
    } else {
        NodeMap<TK, TV>* newNode = new NodeMap<TK, TV>{key, value, map.data[index]};
        map.data[index] = newNode;
        map.length++;
    }

}

// поиск элементов по ключу
template <typename TK, typename TV>
TV GetMap(const MyHashMap<TK, TV>& map, const TK& key) {
    size_t index = HashCode(key, map.capacity);
    NodeMap<TK, TV>* curr = map.data[index];
    while (curr != nullptr) {
        if (curr->key == key) {
            return curr->value;
        }
        curr = curr->next;
    }

    throw;
}


// удаление элементов
template <typename TK, typename TV>
void DeleteMap(MyHashMap<TK, TV>& map, const TK& key) {
    size_t index = HashCode(key, map.capacity);
    NodeMap<TK, TV>* curr = map.data[index];
    NodeMap<TK, TV>* prev = nullptr;
    while (curr != nullptr) {
        if (curr->key == key) {
            if (prev == nullptr) {
                map.data[index] = curr->next;
            } else {
                prev->next = curr->next;
            }
            delete curr;
            map.length--;
            return;
        }
        prev = curr;
        curr = curr->next;
    }
    throw;
}


// очистка памяти
template <typename TK, typename TV>
void DestroyMap(MyHashMap<TK, TV>& map) {
    for (size_t i = 0; i < map.capacity; i++) {
        NodeMap<TK, TV>* curr = map.data[i];
        while (curr != nullptr) {
            NodeMap<TK, TV>* next = curr->next;
            delete curr;
            curr = next;
        }
    }
    delete[] map.data;
    map.data = nullptr;
    map.length = 0;
    map.capacity = 0;
}

template <typename T>
struct MyVector {
    T* data;      //сам массив
    size_t length;        //длина
    size_t capacity;        //capacity - объем
    size_t LoadFactor; //с какого процента заполнения увеличиваем объем = 50%
};

template <typename T>
std::ostream& operator << (std::ostream& os, const MyVector<T>& vector) {
    for (size_t i = 0; i < vector.length; i++) {
        std::cout << vector.data[i];
        if (i < vector.length - 1) std::cout << std::setw(25);
    }
    return os;
}

template <typename T>
MyVector<T>* CreateVector(size_t initCapacity, size_t initLoadFactor) {
    if (initCapacity <= 0 || initLoadFactor <= 0 || initLoadFactor > 100) {
        throw std::runtime_error("Index out of range");
    }

    MyVector<T>* vector = new MyVector<T>;  // Создаем новый вектор
    vector->data = new T[initCapacity];  // Выделяем память под массив
    vector->length = 0;  // Инициализируем длину
    vector->capacity = initCapacity;  // Устанавливаем вместимость
    vector->LoadFactor = initLoadFactor;  // Устанавливаем фактор загрузки
    return vector;
}

// увеличение массива
template <typename T>
void Expansion(MyVector<T>& vector) {
    size_t newCap = vector.capacity * 2;
    T* newData = new T[newCap];
    for (size_t i = 0; i < vector.length; i++) {     //копируем данные из старого массива в новый
        newData[i] = vector.data[i];
    }
    delete[] vector.data;                      // очистка памяти
    vector.data = newData;
    vector.capacity = newCap;
}

// добавление элемента в вектор
template <typename T>
void AddVector(MyVector<T>& vector, T value) {
    if ((vector.length + 1) * 100 / vector.capacity >= vector.LoadFactor) { //обновление размера массива
        Expansion(vector);
    }
    vector.data[vector.length] = value;
    vector.length++;
}


//удаление элемента из вектора
template <typename T>
void DeleteVector(MyVector<T>& vector, size_t index) {
    if (index < 0 || index >= vector.length) {
        throw std::runtime_error("Index out of range");
    }

    for (size_t i = index; i < vector.length - 1; i++) {
        vector.data[i] = vector.data[i + 1];
    }

    vector.length--;
}


// замена элемента по индексу
template <typename T>
void ReplaceVector(MyVector<T>& vector, size_t index, T value) {
    if (index < 0 || index >= vector.length) {
        throw std::runtime_error("Index out of range");
    }
    vector.data[index] = value;
}


struct SchemaInfo {
    string filepath = ".";
    string name;
    int tuplesLimit;
    MyHashMap<string, MyVector<string>*>* jsonStructure;
    map<string, mutex> tableMutexes;
};

enum class NodeType {
    ConditionNode,
    OrNode,
    AndNode
};

// Структура
struct Node {
    NodeType nodeType;
    MyVector<std::string> value;
    Node* left;
    Node* right;

    Node(NodeType type, const MyVector<std::string> val = {}, Node* l = nullptr, Node* r = nullptr)
        : nodeType(type), value(val), left(l), right(r) {}
};


#endif // STRUCTURES_H
