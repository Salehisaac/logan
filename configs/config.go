package configs

import (
    "github.com/joho/godotenv"
    "os"
)

type Config struct {
    LogsPath     string
    ClientSystem string
    AuthKey      string
    StreamFlag   bool
    PhoneNumber string
}


func LoadConfig(flag bool) (*Config, error) {
    err := godotenv.Load()
    if err != nil {
        return nil, err
    }

    return &Config{
        LogsPath:     os.Getenv("LOGS_PATH"),
		StreamFlag: flag,
    }, nil
}