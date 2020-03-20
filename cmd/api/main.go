package main

import (
	"github.com/sergeychur/give_it_away/internal/server"
	"log"
	"os"
)

func main() {
	pathToConfig := ""
	if len(os.Args) != 2 {
		panic("Usage: ./main <path_to_config>")
	} else {
		pathToConfig = os.Args[1]
	}
	serv, err := server.NewServer(pathToConfig)
	if err != nil {
		panic(err)
	}
	err = serv.Run()
	if err != nil {
		log.Println(err)
	}
}
