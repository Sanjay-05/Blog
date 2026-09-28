package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Configuration struct {
	DBHost     string `envconfig:"DB_HOST"`
	DBPort     string `envconfig:"DB_PORT"`
	DBUser     string `envconfig:"DB_USER"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBName     string `envconfig:"DB_NAME"`
	ServerPort string `envconfig:"SERVER_PORT"`
	ApiSecret  string `envconfig:"API_SECRET"`
}

// Loads the .env file assign configuration into the struct
func LoadConfig() (*Configuration, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	cfg := &Configuration{}
	err = envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
