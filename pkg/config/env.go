package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Address                   string
	DBAddr                    string
	JwtExpirationSeconds      int64
	JwtSecret                 string
	AESKey                    string
	MessagesExpirationSeconds int64
}

var Envs = initConfig()

func initConfig() Config {
	godotenv.Load()

	return Config{
		Address:                   getEnv("ADDRESS", "localhost:8080"),
		DBAddr:                    getEnv("DB_ADDR", "oneview.sqlite3"),
		JwtExpirationSeconds:      getEnvAsInt("JWT_EXPIRATION_SECONDS", 3600), // 1 hour
		JwtSecret:                 getEnv("JWT_SECRET", "secret"),
		AESKey:                    getEnv("AES_KEY", "verysecretkey123"),
		MessagesExpirationSeconds: getEnvAsInt("MESSAGES_EXPIRATION_SECONDS", 3600), // 1 hour
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return fallback
		}
		return intValue
	}
	return fallback
}
