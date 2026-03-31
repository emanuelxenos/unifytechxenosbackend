package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort           string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPass               string
	DBName               string
	DBSSLMode            string
	JWTSecret            string
	JWTExpiryHours       int
	RestrictLocalNetwork bool
	AppEnv               string
	AppVersion           string
}

func Load() *Config {
	_ = godotenv.Load()

	expiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "8"))
	restrictLocal, _ := strconv.ParseBool(getEnv("RESTRICT_LOCAL_NETWORK", "false"))

	return &Config{
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPass:               getEnv("DB_PASS", "postgres"),
		DBName:               getEnv("DB_NAME", "mercado_db"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		JWTSecret:            getEnv("JWT_SECRET", "default-secret-change-me"),
		JWTExpiryHours:       expiryHours,
		RestrictLocalNetwork: restrictLocal,
		AppEnv:               getEnv("APP_ENV", "development"),
		AppVersion:           getEnv("APP_VERSION", "1.0.0"),
	}
}

func (c *Config) DatabaseURL() string {
	return "postgres://" + c.DBUser + ":" + c.DBPass + "@" + c.DBHost + ":" + c.DBPort + "/" + c.DBName + "?sslmode=" + c.DBSSLMode
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
