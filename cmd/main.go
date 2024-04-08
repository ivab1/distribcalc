package main

import (
	"net/http"

	"github.com/ivab1/distribcalc/front"
	"github.com/ivab1/distribcalc/internal/agent"
	"github.com/ivab1/distribcalc/internal/database"
	"github.com/ivab1/distribcalc/internal/orchestrator"
)

func main() {
	// Подключение к базе данных
	db := database.StartDB()
	defer db.Close()

	// Создание таблиц
	database.MakeDB(db)

	// Запуск оркестратора
	orchestrator.HandleOrchestrator(db)

	// Запуск клиента
	front.HandlePages(db)

	// Запуск агента
	go agent.Agent(db)

	// Старт сервера
	http.ListenAndServe(":8080", nil)
}
