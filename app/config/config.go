package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Env struct {
	AppName string
	AppEnv  string
	AppPort string
	AppURL  string

	DBDriver string
	DBHost   string
	DBPort   string
	DBName   string
	DBUser   string
	DBPass   string
}

func (config *Env) Init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal("Error loading .env file")
	}

	config.AppName = os.Getenv("APP_NAME")
	config.AppEnv = os.Getenv("APP_ENV")
	config.AppPort = os.Getenv("APP_PORT")
	config.AppURL = os.Getenv("APP_URL")

	config.DBDriver = os.Getenv("DB_DRIVER")
	config.DBHost = os.Getenv("DB_HOST")
	config.DBPort = os.Getenv("DB_PORT")
	config.DBName = os.Getenv("DB_NAME")
	config.DBUser = os.Getenv("DB_USER")
	config.DBPass = os.Getenv("DB_PASS")
}
