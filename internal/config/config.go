package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config contém todas as configurações da aplicação
type Config struct {
	OpenAI   OpenAIConfig   `json:"openai"`
	Database DatabaseConfig `json:"database"`
	App      AppConfig      `json:"app"`
}

// OpenAIConfig contém configurações para integração com OpenAI
type OpenAIConfig struct {
	APIKey string `json:"api_key" env:"OPENAI_API_KEY"`
	Model  string `json:"model" env:"OPENAI_MODEL"`
}

// DatabaseConfig contém configurações para MongoDB
type DatabaseConfig struct {
	URI        string `json:"uri" env:"MONGO_URI"`
	Database   string `json:"database" env:"MONGO_DATABASE"`
	Collection string `json:"collection" env:"MONGO_COLLECTION"`
}

// AppConfig contém configurações gerais da aplicação
type AppConfig struct {
	LogLevel       string        `json:"log_level" env:"LOG_LEVEL"`
	RequestTimeout time.Duration `json:"request_timeout" env:"REQUEST_TIMEOUT"`
	SearchLimit    int           `json:"search_limit" env:"SEARCH_LIMIT"`
	DefaultQuery   string        `json:"default_query" env:"DEFAULT_QUERY"`
}

// Load carrega as configurações a partir das variáveis de ambiente
func Load() (*Config, error) {
	// Tenta carregar arquivo .env (ignora erro se não existir)
	_ = godotenv.Load()
	config := &Config{
		OpenAI: OpenAIConfig{
			APIKey: getEnvOrDefault("OPENAI_API_KEY", ""),
			Model:  getEnvOrDefault("OPENAI_MODEL", "gpt-4-turbo-preview"),
		},
		Database: DatabaseConfig{
			URI:        getEnvOrDefault("MONGO_URI", "mongodb://admin:password123@localhost:27017"),
			Database:   getEnvOrDefault("MONGO_DATABASE", "rag_docs"),
			Collection: getEnvOrDefault("MONGO_COLLECTION", "documents"),
		},
		App: AppConfig{
			LogLevel:       getEnvOrDefault("LOG_LEVEL", "info"),
			RequestTimeout: getEnvDurationOrDefault("REQUEST_TIMEOUT", 30*time.Second),
			SearchLimit:    getEnvIntOrDefault("SEARCH_LIMIT", 5),
			DefaultQuery:   getEnvOrDefault("DEFAULT_QUERY", "What are the documents related to Golang performance?"),
		},
	}

	// Validação básica
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	return config, nil
}

// Validate valida se as configurações obrigatórias estão presentes
func (c *Config) Validate() error {
	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY é obrigatória")
	}

	if c.Database.URI == "" {
		return fmt.Errorf("MONGO_URI é obrigatória")
	}

	if c.App.SearchLimit <= 0 {
		return fmt.Errorf("SEARCH_LIMIT deve ser maior que zero")
	}

	return nil
}

// getEnvOrDefault retorna o valor da variável de ambiente ou um valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault retorna o valor da variável de ambiente como int ou um valor padrão
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDurationOrDefault retorna o valor da variável de ambiente como duration ou um valor padrão
func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
