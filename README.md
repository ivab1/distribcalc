#Распределенный вычислитель арифметических выражений

###Запуск проекта
Для запуска проекта необходимо:
1. Открыть папку проекта в терминале
2. Ввести команду:
```
go mod download
```
3. Открыть Docker
4. Ввести в терминал команду:
```
docker-compose up --build -d
```
5. Во втором терминале открыть папку cmd
6. Ввести команду:
```
go run main.go
```
7. Открыть адрес localhost:8080/home в браузере

###Схема работы поекта
Для взаимодействия с сервером используется графический интерфей, представленный 4 страницами: "глвная страница", "установить время выполнения операций", "список выражений", "состояние серверов" 

После нажатия на кнопку "отправить" на главной странице выражение POST запросом отправляется на оркестратор, где записывается в базу данных и разбивается на мелкие подвыражения, которые при получении GET запроса отправляются на агента, который отправляет полученное выражение вычислителям и записывает ответ в базу данных.

На странице "установить время выполнения операций" можно установить время выполнения разрешенных операций, а так же время жизни сервера. Ограничения записываются в базу данных, из которой их получает агент.

На странице "список выражений" находится список недавних выражений, их статус и ответ (при наличии). В случае, если выражение сожержит символы кроме разрешенных, выражение не обрабатывается, и вместо выражния выводится "Выражение невалидно"

На странице "состояние серверов" находится список серверов и их состояние.

![Схема работы проекта](/docs/scheme.png "Project Scheme")