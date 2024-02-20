# Распределенный вычислитель арифметических выражений

### Запуск проекта
Для запуска проекта необходимо:
1. Открыть папку проекта в терминале
2. Ввести команду для установки зависимостей:
```
go mod download
```
3. Открыть Docker
4. Ввести в терминал команду для старта контейнера с базой данных:
```
docker-compose up --build
```
Чтобы сохранить доступ к текущему терминалу, запустить команду:
```
docker-compose up --build -d
```
5. Во втором терминале открыть папку cmd
6. Ввести команду для запуска проекта:
```
go run main.go
```
7. Открыть адрес localhost:8080/home в браузере

### Схема работы поекта
Для вычислений подходят только целые числа => операция деления - целочисленная

Для взаимодействия с сервером используется графический интерфей, представленный 4 страницами: "глвная страница", "установить время выполнения операций", "список выражений", "состояние серверов" 

После нажатия на кнопку "отправить" на главной странице выражение POST запросом отправляется на оркестратор, где записывается в базу данных и разбивается на мелкие подвыражения, которые при получении GET запроса отправляются на агента, который отправляет полученное выражение вычислителям и записывает ответ в базу данных.

На странице "установить время выполнения операций" можно установить время выполнения разрешенных операций, а так же время жизни сервера (по умолчанию установлено 60 секунд). Ограничения записываются в базу данных, из которой их получает агент.

На странице "список выражений" находится список недавних выражений, их статус и ответ (при наличии). В случае, если выражение сожержит символы кроме разрешенных, выражение не обрабатывается, и вместо выражния выводится "Выражение невалидно"

На странице "состояние серверов" находится список серверов и их состояние.

![Схема работы](/docs/scheme.png "Project Scheme")

### Примеры выражений для проверки:
    2+2
Ответ: 4

    2+2*2
Ответ: 6

    20-10/5
Ответ: 18

### Обратная связь

Если возникнут вопросы, Telegram: @I_ivab