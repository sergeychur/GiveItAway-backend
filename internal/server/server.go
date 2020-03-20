package server

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sergeychur/give_it_away/internal/config"
	"github.com/sergeychur/give_it_away/internal/database"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	router *chi.Mux
	db     *database.DB
	config *config.Config
}

func NewServer(pathToConfig string) (*Server, error) {
	server := new(Server)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	subRouter := chi.NewRouter()
	r.Mount("/api/", subRouter)
	server.router = r

	newConfig, err := config.NewConfig(pathToConfig)
	if err != nil {
		return nil, err
	}
	server.config = newConfig
	dbPort, err := strconv.Atoi(server.config.DBPort)
	if err != nil {
		return nil, err
	}
	db := database.NewDB(server.config.DBUser, server.config.DBPass,
		server.config.DBName, server.config.DBHost, uint16(dbPort))
	server.db = db
	return server, nil
}

func (serv *Server) Run() error {
	err := serv.db.Start()
	if err != nil {
		log.Printf("Failed to connect to DB: %s", err.Error())
		return err
	}
	defer serv.db.Close()
	port := serv.config.Port
	log.SetOutput(os.Stdout)
	log.Printf("Running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, serv.router))
	return nil
}