package service

import (
	"net/http"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"log"
	"strconv"
	"errors"
	"sync"
	"os"
	"os/signal"
	"syscall"
	"net/url"
	"time"
	"sort"
)

const (
	layout        = "2006-01-02"
	filename_conf = "db_conf.json"
)

type dbInfo map[string]string

type Service struct {
	port   int
	srv    *http.Server
	db     *sqlx.DB
	dbinfo dbInfo
}

func NewService(port int) *Service {
	return &Service{port: port, srv: &http.Server{Addr: ":" + strconv.Itoa(port)},
		dbinfo: map[string]string{"engine": "", "username": "", "pass": "", "dbname": "", "port": "5432"}}
}
func (s *Service) Run() (err error) {

	if err = s.parseConfFile(filename_conf); err != nil {
		return err
	}
	if err = s.connectToDB(s.dbinfo); err != nil {
		return err
	}

	s.registerHandleFuncs()

	s.signalProcessing()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//defer wg.Done()
		if err = s.srv.ListenAndServe(); err != nil {
			wg.Done()
		}
	}()

	wg.Wait()
	return nil
}

func (s *Service) connectToDB(info dbInfo) (err error) {
	str := fmt.Sprintf("postgresql://localhost/%s?user=%s&password=%s&port=%s&sslmode=disable",
		s.dbinfo["dbname"], s.dbinfo["username"], s.dbinfo["pass"], s.dbinfo["port"])
	s.db, err = sqlx.Connect(s.dbinfo["engine"], str)
	if err != nil {
		return errors.New("Failed connection to database")
	}
	return nil
}

func (s *Service) parseConfFile(filename string) error {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(data), &s.dbinfo)

	if err != nil {
		return errors.New("Incorrect configuration file")
	}
	return nil
}

func (s *Service) registerUsers(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {

		decoder := json.NewDecoder(req.Body)

		var values map[string]interface{}

		err := decoder.Decode(&values)

		if err != nil {
			s.writeResponse(w, "Incorrect JSON format!\n Try again\n", http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		row, err := s.db.Exec("INSERT INTO users VALUES ($1, $2, $3)", values["id"], values["age"], values["sex"])

		if err != nil {
			s.writeResponse(w, "This ID already exist!", http.StatusInternalServerError)
			return
		}

		rowsAff, err := row.RowsAffected()

		s.writeResponse(w, fmt.Sprintf("Query OK, %d row(s) affected\n", rowsAff), http.StatusOK)

	} else {
		s.writeResponse(w, nil, http.StatusMethodNotAllowed)
		return
	}
}

func (s *Service) getStat(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		params, err := url.ParseQuery(req.URL.RawQuery)

		if err != nil {
			s.writeResponse(w, "Incorrect query rows!", http.StatusBadRequest)
			return
		}

		err = s.validateParams(params)

		if err != nil {
			s.writeResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

		result := map[string][]map[string]interface{}{}

		if err != nil {
			s.writeResponse(w, "Incorrect format date\nPlease, use the format: YYYY-MM-DD\n", http.StatusBadRequest)
			return
		}

		rows, err := s.db.Query(`SELECT time, id, age, sex, count
FROM (
       SELECT
         users.*,
         count(stats.*),
         cast(stats.time AS DATE),
         row_number() OVER (PARTITION BY time ORDER BY count(stats.*) DESC) AS rank
       FROM stats, users
       WHERE cast(time AS DATE) > cast($1 AS DATE) AND
             cast(time AS DATE) < cast($2 AS DATE) AND
             action = $3 AND users.id = stats.user
       GROUP BY users.id, stats.time
       ORDER BY count DESC) t
WHERE rank < $4
ORDER BY time ,rank ASC;`, params["date1"][0], params["date2"][0], params["action"][0], params["limit"][0])

		defer rows.Close()

		if err != nil {
			s.writeResponse(w, nil, http.StatusInternalServerError)
			return
		}

		cols, _ := rows.Columns()

		for rows.Next() {

			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))

			for k := 0; k < len(columns); k++ {
				columnPointers[k] = &columns[k]
			}

			if err := rows.Scan(columnPointers...); err != nil {
				s.writeResponse(w, nil, http.StatusInternalServerError)
			}

			m := make(map[string]interface{})

			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				m[colName] = *val
			}

			date := m["time"].(time.Time).Format(layout)

			result[date] = append(result[date], m)

		}

		rr := []map[string]interface{}{}

		var keys []string

		for k := range result {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			rr = append(rr, map[string]interface{}{"date": k, "rows": result[k]})
		}

		responseJSON := map[string]interface{}{}

		responseJSON["items"] = rr

		data, _ := json.Marshal(responseJSON)

		s.writeResponse(w, string(data)+"\n", http.StatusOK)
	} else {
		s.writeResponse(w, nil, http.StatusMethodNotAllowed)
	}
}

func (s *Service) addStat(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		decoder := json.NewDecoder(req.Body)

		var values map[string]interface{}

		err := decoder.Decode(&values)

		if err != nil {
			s.writeResponse(w, "Incorrect JSON format!\n Try again\n", http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		_, err = s.db.Exec("INSERT INTO stats VALUES ($1, $2, $3)", values["user"], values["action"], values["ts"])

		if err != nil {
			s.writeResponse(w, nil, http.StatusInternalServerError)
			return
		}
	} else {
		s.writeResponse(w, nil, http.StatusMethodNotAllowed)
		return
	}
	return
}

func (s *Service) signalProcessing() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	go s.handler(c)
}

func (s *Service) registerHandleFuncs() {
	http.HandleFunc("/api/users", s.registerUsers)
	http.HandleFunc("/api/users/stats", s.addStat)
	http.HandleFunc("/api/users/stats/top", s.getStat)
}

func (s Service) handler(c chan os.Signal) {
	for {
		<-c
		log.Print("Gracefully stopping...")
		s.db.Close()
		s.srv.Shutdown(nil)
		os.Exit(0)
	}
}

func (s *Service) validateParams(params url.Values) error {

	if params["date1"] == nil || params["date2"] == nil || params["action"] == nil || params["limit"] == nil {
		return errors.New(fmt.Sprintf("Incorrect number of params (have %d, must 4)", len(params)))
	}

	if !s.isValidCategoryStat(string(params["action"][0])) {
		return errors.New(fmt.Sprintf(`Action "%s" isn't support
Please, give the correct action ("login", "logout", "like" or "commentary"`,
			params["action"][0]))
	}

	return nil
}

func (s *Service) isValidCategoryStat(category string) bool {
	switch category {
	case
		"login",
		"like",
		"commentary",
		"logout":
		return true
	}
	return false
}

func (s *Service) writeResponse(w http.ResponseWriter, data interface{}, status int) error {
	w.WriteHeader(status)
	if data != nil {
		_, err := fmt.Fprint(w, data)

		if err != nil {
			return err
		}
	}
	return nil
}
