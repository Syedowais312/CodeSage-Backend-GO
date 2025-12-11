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
	GitHubAppID string
	GitHubAppPrivateKey string
	GitHubWebhookSecret string
	GitHubOAuthClientID string
	GitHubOAuthClientSecret string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(" No .env file found, falling back to system environment variables")
	}
	return &Config{
		Port:        getEnv("PORT", "8080"),
		GitHubToken: os.Getenv("GITHUB_TOKEN"),
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		GeminiKey:   os.Getenv("GEMINI_API_KEY"),
		huggingfaceKey:  os.Getenv("HF_API_KEY"),
		GitHubAppID: os.Getenv("GITHUB_APP_ID"),
		GitHubAppPrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		GitHubOAuthClientID: os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		GitHubOAuthClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
	}
}
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
