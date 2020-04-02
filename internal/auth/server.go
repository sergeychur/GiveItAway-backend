package auth

import (
	"github.com/sergeychur/give_it_away/internal/config"
	"google.golang.org/grpc"
	"log"
	"net"
)

type AuthServerImpl struct {
	ServerConfig *config.Config
	rpcServer    *grpc.Server
	AuthClient   AuthClient
}

func NewServer(pathToConfig string) (*AuthServerImpl, error) {
	server := new(AuthServerImpl)

	newConfig, err := config.NewConfig(pathToConfig)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	server.ServerConfig = newConfig

	server.rpcServer = grpc.NewServer()
	RegisterAuthServer(server.rpcServer, NewAuthManager())
	return server, nil
}

func (server *AuthServerImpl) Run() {
	lis, err := net.Listen("tcp", ":"+server.ServerConfig.AuthPort)
	if err != nil {
		log.Printf("Can`t listen port %s", server.ServerConfig.AuthPort)
		return
	}
	log.Printf("Running AuthMS(grps) on port %s", server.ServerConfig.AuthPort)
	log.Fatal(server.rpcServer.Serve(lis))
}