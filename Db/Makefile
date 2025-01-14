# Имя исполняемого файла
TARGET = Database

# Список исходных файлов .cpp
SRCS = main.cpp

# Список объектных файлов .o, которые будут созданы
OBJS = $(SRCS:.cpp=.o)

# Флаги компиляции
CXXFLAGS = -Wall -Wextra -std=c++17 -pthread

# Компилятор
CXX = g++

# Правило для сборки исполняемого файла
$(TARGET): $(OBJS)
	$(CXX) $(OBJS) -pthread -o $(TARGET)

# Правило для компиляции каждого .cpp файла в .o файл
%.o: %.cpp
	$(CXX) $(CXXFLAGS) -c $< -o $@

# Правило для очистки объектных файлов и исполняемого файла
clean:
	rm -f $(OBJS) $(TARGET)

# Правило для перекомпиляции всего проекта
rebuild: clean $(TARGET)

# Правило по умолчанию
.PHONY: all clean rebuild