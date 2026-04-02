package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	LogLevel          string
	JWTSecret         string
	SupabaseURL       string
	RazorpayKeyID     string
	RazorpayKeySecret string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, loading from environment variables")
	}

	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		RazorpayKeyID:     getEnv("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret: getEnv("RAZORPAY_KEY_SECRET", ""),
	}

	cfg.Validate()

	return cfg
}

func (c *Config) Validate() {
	if c.DatabaseURL == "" {
		log.Fatal("FATAL: DATABASE_URL is not set. Platform storage unavailable.")
	}
	if c.JWTSecret == "" {
		log.Fatal("FATAL: JWT_SECRET is not set. Identity & Access Control disabled.")
	}
	if c.RazorpayKeyID == "" || c.RazorpayKeySecret == "" {
		log.Println("WARNING: Razorpay credentials missing. Payment Settlement will fail.")
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
