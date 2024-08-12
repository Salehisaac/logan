package configs

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	LogsPath     string
	ClientSystem string
	AuthKey      string
	StreamFlag   bool
	PhoneNumber  string
	TimeFlag 	 string
}

func LoadConfig(flag bool, time string) (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		LogsPath:   os.Getenv("LOGS_PATH"),
		StreamFlag: flag,
		TimeFlag: time,
	}, nil
}
