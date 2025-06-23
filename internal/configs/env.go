package configs

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		logrus.Warning("Error loading .env file")
	}
}
