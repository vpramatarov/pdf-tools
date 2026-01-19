package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   int
	MaxUploadSizeMB        int64
	CleanupIntervalMinutes int
	UploadDir              string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system environment variables")
	}

	return &Config{
		Port:                   getEnvAsInt("PORT", 8080),
		MaxUploadSizeMB:        getEnvAsInt64("MAX_FILE_UPLOAD_SIZE", 50),
		CleanupIntervalMinutes: getEnvAsInt("CLEANUP_CRON_INTERVAL", 10),
		UploadDir:              getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsInt64(key string, defaultVal int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultVal
}
