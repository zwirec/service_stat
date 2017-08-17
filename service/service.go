package service

import (
	"encoding/json"
	"errors"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"../requestHandler"
)

const (
	filename_conf = "db_conf.json"
)

type dbInfo map[string]string

type Service struct {
	port   string
	srv    *http.Server
	dbinfo dbInfo
	rH     *requestHandler.RequestHandler
}

func NewService(port string) *Service {
	return &Service{port: port, srv: &http.Server{Addr: ":" + port},
		dbinfo: map[string]string{
			"engine":   "",
			"host":     "",
			"port":     "",
			"username": "",
			"pass":     "",
			"dbname":   "",
		},
	}
}
func (s *Service) Run() (err error) {

	f, err := os.OpenFile("service.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return err
	}

	log.SetOutput(f)

	log.Println("Server running...")

	if err := s.parseConfFile(filename_conf); err != nil {
		return err
	}

	s.rH, err = requestHandler.NewHandler(s.dbinfo)

	if err != nil {
		return err
	}

	s.rH.RegisterHandleFunc()

	s.signalProcessing()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		if err = s.srv.ListenAndServe(); err != nil {
			wg.Done()
		}
	}()

	wg.Wait()
	return err
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

func (s *Service) signalProcessing() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	go s.handler(c)
}

func (s Service) handler(c chan os.Signal) {
	for {
		<-c
		log.Print("Gracefully stopping...")
		s.srv.Shutdown(nil)
		os.Exit(0)
	}
}

func (s *Service) validatePOSTregisterParams(params map[string]interface{}) error {

	if params["id"] == nil || params["age"] == nil || params["sex"] == nil || len(params) != 3 {
		return errors.New(`Missing one or more parameters or parameters invalid (use "id", "age" and "sex")`)
	}
	return nil
}
