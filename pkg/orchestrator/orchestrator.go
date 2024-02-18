package orchestrator

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"

	_ "github.com/lib/pq"
	shuntingYard "github.com/mgenware/go-shunting-yard"
)

type ExpressionStruct struct {
	Expression string `json:"expression"`
	ID         int    `json:"id"`
	TimeAdd    int
	TimeSub    int
	TimeMult   int
	TimeDiv    int
	LifeTime   int
}

type SimpleExpressions struct {
	ID       int    `json:"id"`
	MainID   int    `json:"mainid"`
	Variable string `json:"variable"`
}

func StartDB() *sql.DB {
	connStr := "user=calc password=ZVo2buR4oRA5fEq0fb5o dbname=distribcalc sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func SplitExp(expression []string) [][]string {
	expressions := [][]string{}
	remain := []string{}
	var prev string
	var c int
	if slices.Contains(expression, "/") || slices.Contains(expression, "*") {
		for i := 1; i < len(expression); i += 2 {
			if strings.ContainsAny(expression[i], "*/") {
				c++
				if prev != "" {
					expressions = append(expressions, []string{prev, expression[i], expression[i+1]})
				} else {
					expressions = append(expressions, []string{expression[i-1], expression[i], expression[i+1]})
				}
				prev = fmt.Sprintf("x%d", c)
			} else {
				if prev != "" {
					remain = append(remain, prev)
				} else {
					remain = append(remain, expression[i-1])
				}
				prev = ""
				remain = append(remain, expression[i])
			}
		}
		if prev != "" {
			remain = append(remain, prev)
		} else {
			remain = append(remain, expression[len(expression)-1])
		}
	} else {
		remain = expression
	}
	prev = ""
	if slices.Contains(expression, "-") || slices.Contains(expression, "+") {
		for i := 1; i < len(remain); i += 2 {
			c++
			if prev != "" {
				expressions = append(expressions, []string{prev, remain[i], remain[i+1]})
			} else {
				expressions = append(expressions, []string{remain[i-1], remain[i], remain[i+1]})
			}
			prev = fmt.Sprintf("x%d", c)
		}
	}
	return expressions
}

func Orchestrator(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		newExpression := ExpressionStruct{Expression: string(data)}
		statuscode := 200
		if !strings.ContainsAny(newExpression.Expression, "+-*/") || strings.Contains(newExpression.Expression, "**") || strings.ContainsAny(newExpression.Expression, "!@#$%^&()<>|`~\"'") {
			statuscode = 400
			db := StartDB()
			defer db.Close()
			db.QueryRow("insert into expressions (expression, state) values ($1, $2) returning id", newExpression.Expression, statuscode).Scan(&newExpression.ID)
		} else {
			db := StartDB()
			defer db.Close()
			db.QueryRow("insert into expressions (expression, state) values ($1, $2) returning id", newExpression.Expression, statuscode).Scan(&newExpression.ID)
			row := db.QueryRow("select * from timelimits where id = 1")
			var id int
			err = row.Scan(&id, &newExpression.TimeAdd, &newExpression.TimeSub, &newExpression.TimeMult, &newExpression.TimeDiv, &newExpression.LifeTime)
			if err != nil {
				log.Fatal(err)
			}

			tokens, err := shuntingYard.Scan(newExpression.Expression)
			if err != nil {
				log.Fatal(err)
			}
			expressionData := SplitExp(tokens)
			for i, elem := range expressionData[:len(expressionData)-1] {
				_, err := db.Exec("insert into simpleexpressions (mainexpressionid, variable, n1, n2, operation, processing) values ($1, $2, $3, $4, $5, false)", newExpression.ID, fmt.Sprintf("x%d", i+1), elem[0], elem[2], elem[1])
				if err != nil {
					log.Fatal(err)
				}
			}
			_, err = db.Exec("insert into simpleexpressions (mainexpressionid, variable, n1, n2, operation, processing, last) values ($1, $2, $3, $4, $5, false, true)", newExpression.ID, fmt.Sprintf("x%d", len(expressionData)), expressionData[len(expressionData)-1][0], expressionData[len(expressionData)-1][2], expressionData[len(expressionData)-1][1])
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(expressionData)
		}
	}
	if r.Method == "GET" {
		db := StartDB()
		defer db.Close()
		rows, err := db.Query("select * from simpleexpressions where processing = false")
		if err != nil {
			log.Fatal(err)
		}
		if rows.Next() {
			expr := SimpleExpressions{}
			var a, b, c, d, e, l any
			err = rows.Scan(&expr.ID, &expr.MainID, &expr.Variable, &a, &b, &c, &d, &e, &l)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(expr)
			data, err := json.Marshal(expr)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(data)
		}
	}
}

func HandleOrchestrator() {
	http.HandleFunc("/orchestrator", Orchestrator)
}
