package main

import (
	"../service"
	"log"
	"os"
)

func main() {
	mainLogger := log.New(os.Stderr, "", log.LstdFlags)
	serv := service.NewService("1234")
	if err := serv.Run(); err != nil {
		mainLogger.Println(err)
		os.Exit(1)
	}
}
