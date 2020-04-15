package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	Host           string   `json:"host"`
	Port           string   `json:"port"`
	AuthHost       string   `json:"auth_host"`
	AuthPort       string   `json:"auth_port"`
	DBHost         string   `json:"dbhost"`
	DBPort         string   `json:"dbport"`
	DBUser         string   `json:"dbuser"`
	DBPass         string   `json:"dbpassword"`
	DBName         string   `json:"dbname"`
	Secret         string   `json:"secret"`
	AllowedHosts   []string `json:"allowedHosts,omitempty"`
	UploadPath     string   `json:"uploadPath,omitempty"`
	ApiKey         string   `json:"api_key"`
	CentrifugoPort string   `json:"centrifugo_port"`
	CentrifugoHost string   `json:"centrifugo_host"`
}

func NewConfig(pathToConfig string) (*Config, error) {
	conf := new(Config)
	configFile, err := os.Open(pathToConfig)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer func() {
		_ = configFile.Close()
	}()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&conf)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return conf, nil
}
