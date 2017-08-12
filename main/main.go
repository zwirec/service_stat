package main

import (
	"../service"
	"log"
)

func main() {
	serv := service.NewService(1234)
	if err := serv.Run(); err != nil {
		log.Fatal(err)
	}
}
