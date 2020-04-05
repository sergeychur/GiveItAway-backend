package main

import (
	"github.com/sergeychur/give_it_away/internal/auth"
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

	serv, err := auth.NewServer(pathToConfig)
	if err != nil {
		log.Println(err)
		return
	}
	err = serv.Run()
	if err != nil {
		log.Println(err)
	}
}
