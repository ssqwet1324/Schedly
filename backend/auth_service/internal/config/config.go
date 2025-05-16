package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	DbName     string `env:"DB_NAME"`
	DbUser     string `env:"DB_USER"`
	DbPassword string `env:"DB_PASSWORD"`
	DbHost     string `env:"DB_HOST"`
	DbPort     int    `env:"DB_PORT"`
	JwtSecret  string `env:"JWT_SECRET"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	_ = godotenv.Load()
	cfg.DbName = os.Getenv("DB_NAME")
	cfg.DbUser = os.Getenv("DB_USER")
	cfg.DbPassword = os.Getenv("DB_PASSWORD")
	cfg.DbHost = os.Getenv("DB_HOST")
	cfg.DbPort, _ = strconv.Atoi(os.Getenv("DB_PORT"))
	cfg.JwtSecret = os.Getenv("JWT_SECRET")
	return &cfg, nil
}
