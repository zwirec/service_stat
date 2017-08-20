package requestHandler

import (
	"encoding/json"
	"log"
	"bytes"
	"net/http/httptest"
	"io/ioutil"
	"testing"
	"net/http"
	"github.com/zwirec/http_service_stat/dbManager"
	"github.com/DATA-DOG/go-sqlmock"
	_"github.com/lib/pq"
	"os"
	"fmt"
)

func TestAddUsers(t *testing.T) {

	log.SetFlags(log.Llongfile)
	tests, _ := ioutil.ReadFile("AddUser.json")

	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	JSON := []map[string]interface{}{}

	if err := json.Unmarshal(tests, &JSON); err != nil {
		log.Fatal(err)
	}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	handler := http.HandlerFunc(rH.RegisterUsers)

	for _, request := range JSON {

		var person []byte

		person, err := json.Marshal(request["test"])

		persons_info := map[string]string{}

		err = json.Unmarshal(person, &persons_info)

		_, err = json.Marshal(request["expected"])

		if err != nil {
			log.Fatal(err)
		}

		mock.ExpectExec("INSERT INTO users (.*)").WithArgs(
			persons_info["id"], persons_info["age"], persons_info["sex"]).WillReturnResult(sqlmock.NewResult(1, 1))

		req, err := http.NewRequest("POST", "http://localhost:1234/api/users", bytes.NewBuffer(person))

		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		if err != nil {
			log.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != int(request["expected"].(float64)) {
			t.Fatalf("handler returned wrong status code: got %v want %v",
				status, request["expected"])
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expections: %s", err)
		}
	}
}

func TestAddUserMethodNotAllowed(t *testing.T) {

	db, _, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	req, err := http.NewRequest("GET", "http://localhost:1234/api/users", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	if err != nil {
		log.Fatal(err)
	}

	handler := http.HandlerFunc(rH.RegisterUsers)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)

	}
}

func TestAddUserBadRequest(t *testing.T) {

	db, _, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	body := []string{`{"id::"}`,
		`{
			"id": "2",
			"age": "18",
			"sex": "Gsd"
		}`,
		`{
			"incorrect_field": "2",
			"age": "18",
			"sex": "F"
		}`}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	handler := http.HandlerFunc(rH.RegisterUsers)

	for _, b := range body {
		req, err := http.NewRequest("POST", "http://localhost:1234/api/users", bytes.NewBuffer([]byte(b)))

		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		if err != nil {
			log.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)

		}
	}
}

func TestAddUserIntervalServerError(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	body := []string{
		`{
			"id": "2",
			"age": "18",
			"sex": "M"
		}`}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	handler := http.HandlerFunc(rH.RegisterUsers)

	for _, b := range body {
		req, err := http.NewRequest("POST", "http://localhost:1234/api/users", bytes.NewBuffer([]byte(b)))

		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectExec("INSERT INTO users (.*)").WithArgs("2", "18", "M").WillReturnError(
			fmt.Errorf("smth error"))
		rr := httptest.NewRecorder()

		if err != nil {
			log.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expections: %s", err)
		}
	}
}

func TestAddStatBadRequest(t *testing.T) {
	db, _, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	body := []string{`{"id::"}`,
		`{
			"user": "2",
			"action": "invalid_action",
			"ts": "2012-01-01"
		}`,
		`{
			"incorrect_field": "2",
			"action": "18",
			"ts": "2012-10-10"
		}`}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	handler := http.HandlerFunc(rH.AddStat)

	for _, b := range body {
		req, err := http.NewRequest("POST", "http://localhost:1234/api/users/stats", bytes.NewBuffer([]byte(b)))

		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()

		if err != nil {
			log.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)

		}
	}
}

func TestAddStatInternalServerError(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	body := []string{
		`{
			"user": "2",
			"action": "like",
			"ts": "2012-02-02"
		}`}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	handler := http.HandlerFunc(rH.AddStat)

	for _, b := range body {
		req, err := http.NewRequest("POST", "http://localhost:1234/api/users/stats", bytes.NewBuffer([]byte(b)))

		if err != nil {
			t.Fatal(err)
		}
		mock.ExpectExec("INSERT INTO (.*)").WithArgs("2", "like", "2012-02-02").WillReturnError(
			fmt.Errorf("smth error"))
		rr := httptest.NewRecorder()

		if err != nil {
			log.Fatal(err)
		}

		handler.ServeHTTP(rr, req)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("there were unfulfilled expections: %s", err)
		}
	}
}

func TestAddStatMethodNotAllowed(t *testing.T) {
	db, _, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	req, err := http.NewRequest("GET", "http://localhost:1234/api/users/stats", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	if err != nil {
		log.Fatal(err)
	}

	handler := http.HandlerFunc(rH.AddStat)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)

	}
}

func TestGetStatMethodNotAllowed(t *testing.T) {
	db, _, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	req, err := http.NewRequest("POST", "http://localhost:1234/api/users/stats/top", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	if err != nil {
		log.Fatal(err)
	}

	handler := http.HandlerFunc(rH.GetStat)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)

	}
}

func TestGetIncorrectQueryRow(t *testing.T) {
	db, _, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	handler := http.HandlerFunc(rH.GetStat)

	req, err := http.NewRequest("GET", "http://localhost:1234/api/users/stats/top?fапп a:a1&&", nil)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	if err != nil {
		log.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)

	}

}

func TestGetOK(t *testing.T) {
	db, mock, err := sqlmock.New()

	if err != nil {
		log.Fatal(err)
	}

	dbm := dbManager.DBManager{DB: db}

	rH := RequestHandler{DBManager: &dbm, logger: log.New(os.Stdout, "", log.LstdFlags)}

	handler := http.HandlerFunc(rH.GetStat)

	req, err := http.NewRequest("GET", "http://localhost:1234/api/users/stats/top?date1=2012-02-02&date2=2012-03-10&action=like&limit=1", nil)

	rows := sqlmock.NewRows(nil)

	mock.ExpectQuery("SELECT ").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnRows(rows)

	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	if err != nil {
		log.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)

	}

}
