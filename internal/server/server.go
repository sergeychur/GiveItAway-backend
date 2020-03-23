package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sergeychur/give_it_away/internal/config"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/middlewares"
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
	const idPattern = "^[0-9]+$"
	server := new(Server)
	r := chi.NewRouter()

	newConfig, err := config.NewConfig(pathToConfig)
	if err != nil {
		return nil, err
	}
	server.config = newConfig

	r.Use(middleware.Logger,
		middleware.Recoverer,
		middlewares.CreateCorsMiddleware(server.config.AllowedHosts))

	// upload
	r.Get("/upload/{dir:.+}/{file:.+\\..+$}", http.StripPrefix("/upload/",
		http.FileServer(http.Dir(server.config.UploadPath))).ServeHTTP)

	subRouter := chi.NewRouter()
	// ad
	subRouter.Post("/ad/create", server.CreateAd)
	subRouter.Get(fmt.Sprintf("/ad/{ad_id:%s}/details", idPattern), server.GetAdInfo)
	subRouter.Get("/ad/find", server.FindAds)
	subRouter.Post(fmt.Sprintf("/ad/{ad_id:%s}/upload_image", idPattern), server.AddPhotoToAd)

	// user
	subRouter.Post("/user/auth", server.AuthUser)
	subRouter.Get(fmt.Sprintf("/user/{user_id:%s}", idPattern), server.GetUserInfo)



	r.Mount("/api/", subRouter)


	server.router = r

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
