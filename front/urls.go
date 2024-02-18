package front

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"

	"github.com/ivab1/distribcalc/pkg/orchestrator"
)

type MainExpression struct {
	Id         int
	Expression string
	Answer     any
	State      int
}

type TimeStruct struct {
	Add      int
	Sub      int
	Mult     int
	Div      int
	Lifetime int
}

type Server struct {
	ServerId int
	NextPing string
	State    int
}

func GetExpressionDataFromDB() []MainExpression {
	db := orchestrator.StartDB()
	defer db.Close()
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

func GetSereverInformationFromDB() []Server {
	db := orchestrator.StartDB()
	defer db.Close()
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
	tmpl, err := template.ParseFiles("../front/templates/index.html", "../front/templates/layout.html")
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
	tmpl, err := template.ParseFiles("../front/templates/limits.html", "../front/templates/layout.html")
	if err != nil {
		log.Fatal(err)
	}
	db := orchestrator.StartDB()
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

func ExpressionsPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("../front/templates/expressions.html", "../front/templates/layout.html")
	if err != nil {
		log.Fatal(err)
	}
	data := GetExpressionDataFromDB()
	slices.Reverse(data)
	tmpl.ExecuteTemplate(w, "expressions", data)
}

func StatePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("../front/templates/state.html", "../front/templates/layout.html")
	if err != nil {
		log.Fatal(err)
	}
	serverData := GetSereverInformationFromDB()
	slices.Reverse(serverData)
	tmpl.ExecuteTemplate(w, "state", serverData)
}

func HandlePages() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../front/static"))))
	http.HandleFunc("/home", HomePage)
	http.HandleFunc("/limits", LimitPage)
	http.HandleFunc("/expressions", ExpressionsPage)
	http.HandleFunc("/state", StatePage)
}
