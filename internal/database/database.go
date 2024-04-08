package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/ivab1/distribcalc/internal/authorization"
	_ "github.com/lib/pq"
)

type MainExpression struct {
	Id         int
	Expression string
	Answer     any
	State      int
}

type Server struct {
	ServerId int
	NextPing string
	State    int
}

// Подключение к базе данных
func StartDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	// connStr := "user=calc password=ZVo2buR4oRA5fEq0fb5o dbname=distribcalc sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func MakeDB(db *sql.DB) {
	// // Подключение к базе данных
	// db := StartDB()
	// defer db.Close()

	// Создание таблицы expressions, если такой не существует
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS public.expressions (id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 ), expression text, answer real, state integer, PRIMARY KEY (id));")
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы timelimits, если такой не существует
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.timelimits (id integer DEFAULT 1, add integer DEFAULT 1, sub integer DEFAULT 1, mult integer DEFAULT 1, div integer DEFAULT 1, lifetime integer DEFAULT 1, PRIMARY KEY (id));")
	if err != nil {
		log.Fatal(err)
	}

	// Установка ограничий времени, если не установлены
	_, err = db.Exec("INSERT INTO public.timelimits VALUES (1, 60, 60, 60, 60, 60) ON CONFLICT (id) DO NOTHING")
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы simpleexpressions для хранения подвыражений, если такой не существует
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.simpleexpressions (id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 ), mainexpressionid integer, variable text, n1 text, n2 text, operation char, answer real, processing bool, last bool,  PRIMARY KEY (id));")
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы servers для отслеживания состояния серверов, если такой не существует
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.servers (serverid integer, nextping text, state integer, PRIMARY KEY (serverid));")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO public.servers VALUES (1, $1, 1) ON CONFLICT (serverid) DO NOTHING", time.Now().Add(10*time.Second).Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы с пользователями
	// const usersTable = `
	// CREATE TABLE IF NOT EXISTS users(
	// 	id INTEGER PRIMARY KEY AUTOINCREMENT, 
	// 	name TEXT UNIQUE,
	// 	password TEXT
	// );`

	// if _, err := db.Exec(usersTable); err != nil {
	// 	log.Fatal(err)
	// }
}

// Получение списка выражений
func GetExpressionData(db *sql.DB) []MainExpression {
	// db := StartDB()
	// defer db.Close()
	rows, err := db.Query("select * from expressions")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	mainExpressions := []MainExpression{}
	for rows.Next() {
		mainExpression := MainExpression{}
		err := rows.Scan(&mainExpression.Id, &mainExpression.Expression, &mainExpression.Answer, &mainExpression.State)
		if err != nil {
			log.Fatal(err)
			continue
		}
		mainExpressions = append(mainExpressions, mainExpression)
	}
	return mainExpressions
}

// Получение информации о вычислителях
func GetSereverInfo(db *sql.DB) []Server {
	// db := StartDB()
	// defer db.Close()
	rows, err := db.Query("select * from servers")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	serverInfo := []Server{}
	for rows.Next() {
		server := Server{}
		err := rows.Scan(&server.ServerId, &server.NextPing, &server.State)
		if err != nil {
			log.Fatal(err)
			continue
		}
		serverInfo = append(serverInfo, server)
	}
	return serverInfo
}

// Добавление нового пользователя в таблицу
func InsertUser(ctx context.Context, db *sql.DB, user *authorization.User) (int64, error) {
	var q = `
	INSERT INTO users (name, password) values ($1, $2)
	`
	result, err := db.ExecContext(ctx, q, user.Name, user.Password)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Получение пользователя по имени
func SelectUser(ctx context.Context, db *sql.DB, name string) (authorization.User, error) {
	var (
		user authorization.User
		err  error
	)

	var q = "SELECT id, name, password FROM users WHERE name=$1"
	err = db.QueryRowContext(ctx, q, name).Scan(&user.ID, &user.Name, &user.Password)
	return user, err
}