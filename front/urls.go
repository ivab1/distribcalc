package front

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/ivab1/distribcalc/internal/authorization"
	"github.com/ivab1/distribcalc/internal/database"
	"github.com/ivab1/distribcalc/internal/orchestrator"
)

type TimeStruct struct {
	Add      int
	Sub      int
	Mult     int
	Div      int
	Lifetime int
}

func SendoToOrchestrator(exp orchestrator.ExpressionStruct) {
	data, err := json.Marshal(exp)
	if err != nil {
		log.Fatal(err)
	}
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
		cookieData, err := r.Cookie("user")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			token := cookieData.Value
			user := authorization.GetTokenValue(token)
			expression := orchestrator.ExpressionStruct{Expression: r.FormValue("expression"), UserID: int(user.ID)}
			SendoToOrchestrator(expression)
		}
	}
	tmpl.ExecuteTemplate(w, "index", "")
}

func LimitPage(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./front/templates/limits.html", "./front/templates/layout.html")
		if err != nil {
			log.Fatal(err)
		}
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
	})
}

func ExpressionsPage(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("./front/templates/expressions.html", "./front/templates/layout.html")
		if err != nil {
			log.Fatal(err)
		}
		cookieData, err := r.Cookie("user")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			token := cookieData.Value
			user := authorization.GetTokenValue(token)
			data := database.GetExpressionData(db, user.ID)
			slices.Reverse(data)
			tmpl.ExecuteTemplate(w, "expressions", data)
		}
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

func RegistrationPage(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var info string
		tmpl, err := template.ParseFiles("./front/templates/registration.html")
		if err != nil {
			log.Fatal(err)
		}
		if r.Method == "POST" {
			username := r.FormValue("username")
			password1 := r.FormValue("password1")
			password2 := r.FormValue("password2")
			if password1 != password2 {
				log.Println("Пароли не совпадают")
				info = "Пароли не совпадают!"
			} else {
				user, err := authorization.MakeUser(username, password1)
				if err != nil {
					info = "Ошибка регистрации!"
				} else {
					id, err := database.InsertUser(context.Background(), db, &user)
					if err != nil {
						info = "Ошибка регистрации!"
					} else {
						log.Printf("New user: %s id: %d!", user.Name, id)
						user.ID = id
						token := authorization.MakeToken(user)
						http.SetCookie(w, &http.Cookie{
							Name:     "user",
							Value:    token,
							HttpOnly: true,
							Expires: time.Now().Add(5 * time.Minute),
						})
						http.Redirect(w, r, "/home", http.StatusFound)
					}
				}
			}
		}
		tmpl.Execute(w, info)
	})
}

func LoginPage(db *sql.DB) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var info string
		tmpl, err := template.ParseFiles("./front/templates/login.html")
		if err != nil {
			log.Fatal(err)
		}
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")
			user, err := database.SelectUser(context.TODO(), db, username)
			if err != nil {
				info = "Ошибка входа!"
			} else {
				err = authorization.User{Name: username, OriginPassword: password}.ComparePassword(user)
				if err != nil {
					info = "Ошибка входа!"
				} else {
					token := authorization.MakeToken(user)
					http.SetCookie(w, &http.Cookie{
						Name:     "user",
						Value:    token,
						HttpOnly: true,
						Expires: time.Now().Add(5 * time.Minute),
					})
					http.Redirect(w, r, "/home", http.StatusFound)
				}
			}
		}
		tmpl.Execute(w, info)
	})
}

func StartPage(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("user")
	if err != nil {
		tmpl, err := template.ParseFiles("./front/templates/start.html")
		if err != nil {
			log.Fatal(err)
		}
		tmpl.Execute(w, "")
	} else {
		http.Redirect(w, r, "/home", http.StatusFound)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user",
		Value:    "",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func HandlePages(db *sql.DB) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./front/static"))))
	http.HandleFunc("/home", HomePage)
	http.HandleFunc("/", StartPage)
	http.HandleFunc("/limits", LimitPage(db))
	http.HandleFunc("/expressions", ExpressionsPage(db))
	http.HandleFunc("/state", StatePage(db))
	http.HandleFunc("/registration", RegistrationPage(db))
	http.HandleFunc("/login", LoginPage(db))
	http.HandleFunc("/logout", Logout)
}
