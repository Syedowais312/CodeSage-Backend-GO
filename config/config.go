package config

import (
	"os"
	"log"
 	 "github.com/joho/godotenv"
)

type Config struct {
	Port        string
	GitHubToken string
	OpenAIKey   string
	GeminiKey   string
	huggingfaceKey string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, falling back to system environment variables")
	}
	return &Config{
		Port:        getEnv("PORT", "8080"),
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		GeminiKey:   os.Getenv("GEMINI_API_KEY"),
		huggingfaceKey:  os.Getenv("HF_API_KEY"),
	}
}
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
