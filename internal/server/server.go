package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sergeychur/give_it_away/internal/auth"
	"github.com/sergeychur/give_it_away/internal/config"
	"github.com/sergeychur/give_it_away/internal/database"
	"github.com/sergeychur/give_it_away/internal/middlewares"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	router *chi.Mux
	db     *database.DB
	config *config.Config
	AuthClient auth.AuthClient
	CookieField string
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
	server.CookieField = "token"

	r.Use(middleware.Logger,
		middleware.Recoverer,
		middlewares.CreateCorsMiddleware(server.config.AllowedHosts))
	needLogin := chi.NewRouter()
	needLogin.Use(middlewares.CreateCheckAuthMiddleware([]byte(server.config.Secret), server.CookieField, server.IsLogined))
	// upload
	r.Get("/upload/{dir:.+}/{file:.+\\..+$}", http.StripPrefix("/upload/",
		http.FileServer(http.Dir(server.config.UploadPath))).ServeHTTP)

	subRouter := chi.NewRouter()
	// ad
	needLogin.Post("/ad/create", server.CreateAd)
	needLogin.Put(fmt.Sprintf("/ad/{ad_id:%s}/edit", idPattern), server.EditAd)
	subRouter.Get(fmt.Sprintf("/ad/{ad_id:%s}/details", idPattern), server.GetAdInfo)
	needLogin.Get("/ad/find", server.FindAds)
	needLogin.Post(fmt.Sprintf("/ad/{ad_id:%s}/upload_image", idPattern), server.AddPhotoToAd)
	needLogin.Post(fmt.Sprintf("/ad/{ad_id:%s}/delete", idPattern), server.DeleteAd)
	needLogin.Post(fmt.Sprintf("/ad/{ad_id:%s}/delete_photo", idPattern), server.DeleteAdPhoto)
	needLogin.Post(fmt.Sprintf("/ad/{ad_id:%s}/set_hidden", idPattern), server.SetHidden)

	// deal
	needLogin.Post(fmt.Sprintf("/ad/{ad_id:%s}/subscribe", idPattern), server.SubscribeToAd)
	subRouter.Get(fmt.Sprintf("/ad/{ad_id:%s}/subscribers", idPattern), server.GetAdSubscribers)		// think about it
	needLogin.Post(fmt.Sprintf("/ad/{ad_id:%s}/unsubscribe", idPattern), server.UnsubscribeFromAd)
	needLogin.Put(fmt.Sprintf("/ad/{ad_id:%s}/make_deal", idPattern), server.MakeDeal)
	needLogin.Get(fmt.Sprintf("/ad/{ad_id:%s}/deal", idPattern), server.CancelDeal)

	needLogin.Post(fmt.Sprintf("/deal/{deal_id:%s}/fulfill", idPattern), server.FulfillDeal)
	needLogin.Post(fmt.Sprintf("/deal/{deal_id:%s}/cancel", idPattern), server.CancelDeal)
	subRouter.Get(fmt.Sprintf("/ad/{ad_id:%s}/deal", idPattern), server.GetDealForAd)

	// notifications
	needLogin.Get("/notifications", server.GetNotifications)

	// user
	subRouter.Post("/user/auth", server.AuthUser)
	subRouter.Get(fmt.Sprintf("/user/{user_id:%s}", idPattern), server.GetUserInfo)

	r.Mount("/api/", subRouter)
	subRouter.Mount("/", needLogin)

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

func (server *Server) Run() error {
	err := server.db.Start()
	if err != nil {
		log.Printf("Failed to connect to DB: %s", err.Error())
		return err
	}
	defer server.db.Close()
	port := server.config.Port
	log.SetOutput(os.Stdout)

	log.Printf("Running on port %s\n", port)
	grcpAuthConn, err := grpc.Dial(
		server.config.AuthHost+":"+server.config.AuthPort,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Println("Can`t connect ro grpc (auth ms)")
	}
	defer func() {
		_ = grcpAuthConn.Close()
	}()

	server.AuthClient = auth.NewAuthClient(grcpAuthConn)


	log.Fatal(http.ListenAndServe(":"+port, server.router))
	return nil
}
