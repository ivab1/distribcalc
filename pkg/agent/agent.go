package agent

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ivab1/distribcalc/pkg/orchestrator"
)

type Server struct {
	ID          int
	NextPing    string
	ServerState int
}

type TimeLimitStruct struct {
	TimeAdd  int
	TimeSub  int
	TimeMult int
	TimeDiv  int
	LifeTime int
}

type ExpressionForCount struct {
	ID         int
	MainID     int
	Variable   string
	N1         string
	N2         string
	Operation  string
	Answer     any
	Processing bool
	Last       bool
}

func Agent() {
	db := orchestrator.StartDB()
	db.Exec("update servers set nextping = $1 where id = 1", time.Now().Add(10*time.Second).Format("2006-01-02 15:04:05"))
	go SendPing(db)
	for {
		time.Sleep(2 * time.Second)
		row := db.QueryRow("select * from timelimits where id = 1")
		limits := TimeLimitStruct{}
		var id int
		err := row.Scan(&id, &limits.TimeAdd, &limits.TimeSub, &limits.TimeMult, &limits.TimeDiv, &limits.LifeTime)
		if err != nil {
			log.Fatal(err)
		}
		StartCountingAndGetResult(limits, db)

		row = db.QueryRow("select * from servers where serverid = 1")
		server := Server{}
		row.Scan(&server.ID, &server.NextPing, &server.ServerState)
		if t, _ := time.Parse("2006-01-02 15:04:05", server.NextPing); time.Until(t) <= time.Duration(limits.LifeTime/4)*time.Second {
			db.Exec("update servers set nextping = $1 where id = 1", time.Now().Add(time.Duration(limits.LifeTime)*time.Second).Format("2006-01-02 15:04:05"))
		}
	}
}

func SendPing(db *sql.DB) {
	for {
		row := db.QueryRow("select * from servers where serverid = 1")
		server := Server{}
		row.Scan(&server.ID, &server.NextPing, &server.ServerState)
		row = db.QueryRow("select * from timelimits where id = 1")
		limits := TimeLimitStruct{}
		var id int
		err := row.Scan(&id, &limits.TimeAdd, &limits.TimeSub, &limits.TimeMult, &limits.TimeDiv, &limits.LifeTime)
		if err != nil {
			log.Fatal(err)
		}
		time1, err := time.Parse("2006-01-02 15:04:05", time.Now().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Fatal(err)
		}
		time2, err := time.Parse("2006-01-02 15:04:05", server.NextPing)
		if err != nil {
			log.Fatal(err)
		}
		if time2.Before(time1) {
			db.Exec("update servers set state = 3, nextping = $1 where id = $2", server.NextPing, id)
		}
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(limits.LifeTime) * time.Second)
	}
}

func StartCountingAndGetResult(limits TimeLimitStruct, db *sql.DB) {
	answer := make(chan int)
	flag, toCount, expId, mainExpId, expVar, lastCheck := GetExpression()
	if flag {
		go Counter(toCount, limits, answer)
		expressionAnswer := <-answer
		close(answer)
		_, err := db.Exec("update simpleexpressions set answer = $1 where id = $2", expressionAnswer, expId)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("update simpleexpressions set n1 = $1 where (mainexpressionid = $2) and (n1 = $3)", expressionAnswer, mainExpId, expVar)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("update simpleexpressions set n2 = $1 where (mainexpressionid = $2) and (n2 = $3)", expressionAnswer, mainExpId, expVar)
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec("update simpleexpressions set processing = true where id = $1", expId)
		if err != nil {
			log.Fatal(err)
		}
		if lastCheck {
			_, err := db.Exec("update expressions set answer = $1, state = 201 where id = $2", expressionAnswer, mainExpId)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func Counter(expression []string, operationTime TimeLimitStruct, answerCh chan int) {
	n1, err := strconv.Atoi(expression[0])
	if err != nil {
		log.Fatal(err)
	}
	n2, err := strconv.Atoi(expression[2])
	if err != nil {
		log.Fatal(err)
	}
	op := expression[1]
	if op == "+" {
		go func() {
			time.Sleep(time.Duration(operationTime.TimeAdd) * time.Second)
			answerCh <- n1 + n2
		}()
	} else if op == "-" {
		go func() {
			time.Sleep(time.Duration(operationTime.TimeSub) * time.Second)
			answerCh <- n1 - n2
		}()
	} else if op == "*" {
		go func() {
			time.Sleep(time.Duration(operationTime.TimeMult) * time.Second)
			answerCh <- n1 * n2
		}()
	} else if op == "/" {
		go func() {
			time.Sleep(time.Duration(operationTime.TimeDiv) * time.Second)
			answerCh <- n1 / n2
		}()
	}
}

func GetExpression() (bool, []string, int, int, string, bool) {
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/orchestrator", nil)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if len(body) > 0 {
		receivedExpression := orchestrator.SimpleExpressions{}
		err = json.Unmarshal(body, &receivedExpression)
		if err != nil {
			log.Fatal(err)
		}
		db := orchestrator.StartDB()
		defer db.Close()
		row := db.QueryRow("select * from simpleexpressions where id = $1", receivedExpression.ID)
		expression := ExpressionForCount{}
		row.Scan(&expression.ID, &expression.MainID, &expression.Variable, &expression.N1, &expression.N2, &expression.Operation, &expression.Answer, &expression.Processing, &expression.Last)
		return true, []string{expression.N1, expression.Operation, expression.N2}, expression.ID, expression.MainID, expression.Variable, expression.Last
	}
	return false, nil, 0, 0, "", false
}
