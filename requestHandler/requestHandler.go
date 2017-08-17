package requestHandler

import (
	"net/http"
	"encoding/json"
	"fmt"
	"sort"
	"net/url"
	"../dbManager"
	"log"
	"errors"
	"time"
	"os"
)

const (
	layout = "2006-01-02"
)

type RequestHandler struct {
	dbmanager *dbManager.DBManager
	logger    *log.Logger
}

func NewHandler(dbinfo map[string]string, logger ... *log.Logger) (*RequestHandler, error) {
	dbm, err := dbManager.NewDBManager(dbinfo)
	if err != nil {
		return nil, err
	}
	r := &RequestHandler{}
	r.dbmanager = dbm

	if logger == nil {
		r.logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		r.logger = logger[0]
	}

	return r, nil
}

func (reqHandler *RequestHandler) RegisterHandleFunc() error {
	http.HandleFunc("/api/users", reqHandler.registerUsers)
	http.HandleFunc("/api/users/stats", reqHandler.addStat)
	http.HandleFunc("/api/users/stats/top", reqHandler.getStat)
	return nil
}

func (reqHandler *RequestHandler) addStat(w http.ResponseWriter, req *http.Request) {
	var httpStatus int
	if req.Method == "POST" {

		decoder := json.NewDecoder(req.Body)

		var values map[string]interface{}

		if err := decoder.Decode(&values); err != nil {
			httpStatus = http.StatusBadRequest
			reqHandler.writeResponse(w, "Incorrect JSON format!\n Try again\n", httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}

		defer req.Body.Close()

		_, err := reqHandler.dbmanager.PutStats(values)

		if err != nil {
			httpStatus = http.StatusInternalServerError
			reqHandler.writeResponse(w, nil, httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}


		if err != nil {
			httpStatus = http.StatusInternalServerError
			reqHandler.writeResponse(w, nil, httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}
	} else {
		httpStatus = http.StatusMethodNotAllowed
		reqHandler.writeResponse(w, nil, httpStatus)
		reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
		return
	}
	return
}

func (reqHandler *RequestHandler) registerUsers(w http.ResponseWriter, req *http.Request) {
	var httpStatus int

	defer reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
	if req.Method == "POST" {

		decoder := json.NewDecoder(req.Body)

		var values map[string]interface{}

		err := decoder.Decode(&values)

		if err != nil {
			httpStatus = http.StatusBadRequest
			reqHandler.writeResponse(w, "Incorrect JSON format!\n Try again\n", httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}

		if err := reqHandler.validatePOSTregisterParams(values); err != nil {
			httpStatus = http.StatusBadRequest
			reqHandler.writeResponse(w, err, httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}

		if !reqHandler.isValidSex(values["sex"].(string)) {
			httpStatus = http.StatusBadRequest
			reqHandler.writeResponse(w, nil, httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}

		defer req.Body.Close()

		_, err = reqHandler.dbmanager.CreateUser(values)

		if err != nil {
			httpStatus = http.StatusBadRequest
			reqHandler.writeResponse(w, err.Error(), httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}

		httpStatus = http.StatusOK
		reqHandler.writeResponse(w, nil, httpStatus)
		reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
	} else {
		httpStatus = http.StatusMethodNotAllowed
		reqHandler.writeResponse(w, nil, httpStatus)
		reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
		return
	}
}

func (reqHanlder *RequestHandler) validatePOSTregisterParams(params map[string]interface{}) error {

	if params["id"] == nil || params["age"] == nil || params["sex"] == nil || len(params) != 3 {
		return errors.New(`Missing one or more parameters or parameters invalid (use "id", "age" and "sex")`)
	}
	return nil
}

func (reqHandler *RequestHandler) getStat(w http.ResponseWriter, req *http.Request) {
	var httpStatus int

	//defer reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)

	if req.Method == "GET" {

		values, err := url.ParseQuery(req.URL.RawQuery)

		if err != nil {
			httpStatus = http.StatusBadRequest
			reqHandler.writeResponse(w, "Incorrect query rows!", httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}

		err = reqHandler.validateGETParams(values)

		if err != nil {
			httpStatus = http.StatusBadRequest
			reqHandler.writeResponse(w, err.Error(), httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			return
		}

		result := map[string][]map[string]interface{}{}

		rows, err := reqHandler.dbmanager.GetStats(values)

		if err != nil {
			httpStatus = http.StatusInternalServerError
			reqHandler.writeResponse(w, nil, httpStatus)
			reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
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
				httpStatus = http.StatusInternalServerError
				reqHandler.writeResponse(w, nil, httpStatus)
				reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
			}

			m := make(map[string]interface{})

			for i, colName := range cols {
				val := columnPointers[i].(*interface{})
				m[colName] = *val
			}

			date := m["date"].(time.Time).Format(layout)

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

		httpStatus = http.StatusOK

		reqHandler.writeResponse(w, string(data)+"\n", httpStatus)
		reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
	} else {
		httpStatus = http.StatusOK
		reqHandler.writeResponse(w, nil, httpStatus)
		reqHandler.logger.Printf(`%s "%s %s %s %d"`, req.RemoteAddr, req.Method, req.URL.Path, req.Proto, httpStatus)
	}
}

func (reqHandler *RequestHandler) validateGETParams(params url.Values) error {

	if params["date1"] == nil || params["date2"] == nil || params["action"] == nil || params["limit"] == nil {
		return fmt.Errorf("Incorrect number of params (have %d, must 4)", len(params))
	}

	if !reqHandler.isValidStatCategory(params["action"][0]) {
		return fmt.Errorf("Incorrect value(s)")
	}

	return nil
}

func (reqHandler *RequestHandler) writeResponse(w http.ResponseWriter, data interface{}, status int) error {
	w.WriteHeader(status)
	if data != nil {
		_, err := fmt.Fprint(w, data)

		if err != nil {
			return err
		}
	}
	return nil
}

func (reqHandler *RequestHandler) isValidStatCategory(category string) bool {
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

func (reqHandler *RequestHandler) isValidSex(sex string) bool {
	if !(sex == "M" || sex == "F") {
		return false
	}
	return true
}
