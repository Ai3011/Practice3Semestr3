#ifndef Utilities_H
#define Utilities_H


#include <iostream>
#include <string>
#include "structures.h"

using namespace std;


// Возвращает подстроку от start до end (не включая end)
string getSubsting(const string &str, int start, int end) {
    string result;
    for (int i = start; i < end; i++) {
        result += str[i];
    }
    return result;
}

// Возвращает длину строки
int getLen(const string &str) {
    int len = 0;
    while (str[len]!= '\0') {
        len++;
    }
    return len;
}

MyVector<string>* splitRow(const string &str, char delim) {
    int index = 0;
    MyVector<string>* slovs = CreateVector<string>(10, 50);
    int length = getLen(str);

    while (true) {
        int delimIndex = index;
        while (str[delimIndex] != delim && delimIndex != length) delimIndex++;

        string word = getSubsting(str, index, delimIndex);
        AddVector(*slovs, word);
        index = delimIndex + 1;
        if (delimIndex == length) break;
    }

    return slovs;
}

#endif 