package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ivab1/distribcalc/front"
	"github.com/ivab1/distribcalc/pkg/agent"
	"github.com/ivab1/distribcalc/pkg/orchestrator"
)

func MakeDB() {
	// Подключение к базе данных
	db := orchestrator.StartDB()
	defer db.Close()

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

	// Создание таблицы servers, если такой не существует
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.servers (serverid integer, nextping text, state integer, PRIMARY KEY (serverid));")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO public.servers VALUES (1, $1, 1) ON CONFLICT (serverid) DO NOTHING", time.Now().Add(10*time.Second).Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Создание таблиц
	MakeDB()

	// Запуск оркестратора
	orchestrator.HandleOrchestrator()

	// Запуск клиента
	front.HandlePages()

	// Запуск агента
	go agent.Agent()

	// Старт сервера
	http.ListenAndServe(":8080", nil)
}
