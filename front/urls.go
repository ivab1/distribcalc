package front

import (
	"bytes"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/ivab1/distribcalc/internal/database"
)



type TimeStruct struct {
	Add      int
	Sub      int
	Mult     int
	Div      int
	Lifetime int
}

func SendoToOrchestrator(exp string) {
	data := []byte(exp)
	req, err := http.NewRequest("POST", "http://localhost:8080/orchestrator", bytes.NewBuffer(data))
	req.Header.Set("Data-Type", "expression")
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./front/templates/index.html", "./front/templates/layout.html")
	if err != nil {
		log.Fatal(err)
	}
	if r.Method == "POST" {
		expression := r.FormValue("expression")
		SendoToOrchestrator(expression)
	}
	tmpl.ExecuteTemplate(w, "index", "")
}

func LimitPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./front/templates/limits.html", "./front/templates/layout.html")
	if err != nil {
		log.Fatal(err)
	}
	db := database.StartDB()
	defer db.Close()
	dataToSend := TimeStruct{}
	if r.Method == "POST" {
		add, _ := strconv.Atoi(r.FormValue("add"))
		sub, _ := strconv.Atoi(r.FormValue("sub"))
		mult, _ := strconv.Atoi(r.FormValue("mult"))
		div, _ := strconv.Atoi(r.FormValue("div"))
		lifetime, _ := strconv.Atoi(r.FormValue("lifetime"))
		_, err1 := db.Exec("update timelimits set add = $1, sub = $2, mult = $3, div = $4, lifetime = $5 where id = 1", add, sub, mult, div, lifetime)
		if err1 != nil {
			log.Fatal(err1)
		}
		dataToSend = TimeStruct{Add: add, Sub: sub, Mult: mult, Div: div, Lifetime: lifetime}
	} else {
		row := db.QueryRow("select * from timelimits where id = 1")
		var a int
		err = row.Scan(&a, &dataToSend.Add, &dataToSend.Sub, &dataToSend.Mult, &dataToSend.Div, &dataToSend.Lifetime)
		if err != nil {
			panic(err)
		}
	}
	tmpl.ExecuteTemplate(w, "limits", dataToSend)
}

func ExpressionsPage(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./front/templates/expressions.html", "./front/templates/layout.html")
		if err != nil {
			log.Fatal(err)
		}
		data := database.GetExpressionData(db)
		slices.Reverse(data)
		tmpl.ExecuteTemplate(w, "expressions", data)
	})
}

func StatePage(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./front/templates/state.html", "./front/templates/layout.html")
		if err != nil {
			log.Fatal(err)
		}
		serverData := database.GetSereverInfo(db)
		slices.Reverse(serverData)
		tmpl.ExecuteTemplate(w, "state", serverData)
	})
}

func HandlePages(db *sql.DB) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./front/static"))))
	http.HandleFunc("/home", HomePage)
	http.HandleFunc("/", HomePage)
	http.HandleFunc("/limits", LimitPage)
	http.HandleFunc("/expressions", ExpressionsPage(db))
	http.HandleFunc("/state", StatePage(db))
}
