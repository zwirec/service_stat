package service

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

func TestAddUsers(t *testing.T) {

	file, _ := ioutil.ReadFile("AddUser.json")

	decoder := json.NewDecoder(bytes.NewBuffer(file))

	JSON := []map[string]interface{}{}

	err := decoder.Decode(&JSON)

	if err != nil {
		log.Fatal(err)
	}

	for _, request := range JSON {

		var person []byte

		person, err := json.Marshal(request["test"])

		expected, err := json.Marshal(request["expected"])

		if err != nil {
			log.Fatal("Incorect JSON")
		}

		req, _ := http.NewRequest("POST", "http://localhost:1234/api/users", bytes.NewBuffer(person))

		client := &http.Client{}

		resp, err := client.Do(req)

		exp, _ := strconv.Atoi(string(expected))

		if resp.StatusCode != exp {
			t.Errorf("Expected %s, got %d", expected, resp.StatusCode)
		}

	}

}

func TestAddStat(t *testing.T) {
	file, _ := ioutil.ReadFile("AddStat.json")

	decoder := json.NewDecoder(bytes.NewBuffer(file))

	JSON := []map[string]interface{}{}

	err := decoder.Decode(&JSON)

	if err != nil {
		log.Fatal(err)
	}

	for _, request := range JSON {

		var stats []byte

		stats, err := json.Marshal(request["test"])

		expected, err := json.Marshal(request["expected"])

		if err != nil {
			log.Fatal("Incorect JSON")
		}

		req, _ := http.NewRequest("POST", "http://localhost:1234/api/users/stats", bytes.NewBuffer(stats))

		client := &http.Client{}

		resp, err := client.Do(req)

		exp, _ := strconv.Atoi(string(expected))

		if resp.StatusCode != exp {
			t.Errorf("Expected %s, got %d", expected, resp.StatusCode)
		}

	}
	return
}

func TestGetStat(t *testing.T) {
	file, _ := ioutil.ReadFile("GetStat.json")

	decoder := json.NewDecoder(bytes.NewBuffer(file))

	JSON := []map[string]interface{}{}

	err := decoder.Decode(&JSON)

	if err != nil {
		log.Fatal(err)
	}

	for _, request := range JSON {

		params := url.Values{}

		test := request["test"].(map[string]interface{})

		params["date1"] = []string{test["date1"].(string)}
		params["date2"] = []string{test["date2"].(string)}
		params["action"] = []string{test["action"].(string)}
		params["limit"] = []string{test["limit"].(string)}

		expected, err := json.Marshal(request["expected"])

		if err != nil {
			log.Fatal("Incorect JSON")
		}

		req, _ := http.NewRequest("GET", "http://localhost:1234/api/users/stats/top?"+params.Encode(), nil)

		client := &http.Client{}

		resp, err := client.Do(req)

		exp, _ := strconv.Atoi(string(expected))

		if resp.StatusCode != exp {
			t.Errorf("Expected %s, got %d", expected, resp.StatusCode)
		}

	}
	return
}
