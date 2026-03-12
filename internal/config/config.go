package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI     string
	Port         string
	DBName       string
	JWTSecret    string
	FrontendURL  string
	AdminUserIDs []string
}

var cfg *Config

func Load() *Config {
	godotenv.Load()

	adminIDs := []string{}
	if raw := getEnv("ADMIN_USER_IDS", ""); raw != "" {
		for _, id := range strings.Split(raw, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				adminIDs = append(adminIDs, id)
			}
		}
	}

	cfg = &Config{
		MongoURI:     getEnv("MONGO_URI", "mongodb://localhost:27017"),
		Port:         getEnv("PORT", "8090"),
		DBName:       getEnv("DB_NAME", "tron_3d"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
		FrontendURL:  getEnv("FRONTEND_URL", "https://whodo.com.br"),
		AdminUserIDs: adminIDs,
	}

	return cfg
}

func Get() *Config {
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
