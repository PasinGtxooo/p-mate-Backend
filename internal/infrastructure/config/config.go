package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	FrontendURL string `mapstructure:"FRONTEND_URL"`
	DBHost      string `mapstructure:"DB_HOST"`
	DBName      string `mapstructure:"DB_DATABASE"`
	DBUsername  string `mapstructure:"DB_USERNAME"`
	DBPassword  string `mapstructure:"DB_PASSWORD"`
	DBPort      string `mapstructure:"DB_PORT"`
	GroqAPIKey  string `mapstructure:"GROQ_API_KEY"`
	GroqAPIURL  string `mapstructure:"GROQ_API_URL"`
}

func NewConfig() *Config {
	config := &Config{}

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalln("Unable to read .env file", err)
		}
	}

	viper.AutomaticEnv()

	if err := viper.BindEnv("FRONTEND_URL"); err != nil {
		log.Fatalf("Unable to bind FRONTEND_URL: %v", err)
	}
	if err := viper.BindEnv("DB_HOST"); err != nil {
		log.Fatalf("Unable to bind DB_HOST: %v", err)
	}
	if err := viper.BindEnv("DB_DATABASE"); err != nil {
		log.Fatalf("Unable to bind DB_DATABASE: %v", err)
	}
	if err := viper.BindEnv("DB_USERNAME"); err != nil {
		log.Fatalf("Unable to bind DB_USERNAME: %v", err)
	}
	if err := viper.BindEnv("DB_PASSWORD"); err != nil {
		log.Fatalf("Unable to bind DB_PASSWORD: %v", err)
	}
	if err := viper.BindEnv("DB_PORT"); err != nil {
		log.Fatalf("Unable to bind DB_PORT: %v", err)
	}
	if err := viper.BindEnv("GROQ_API_KEY"); err != nil {
		log.Fatalf("Unable to bind GROQ_API_KEY: %v", err)
	}
	if err := viper.BindEnv("GROQ_API_URL"); err != nil {
		log.Fatalf("Unable to bind GROQ_API_URL: %v", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		log.Fatalln("Unable to decode into struct", err)
	}

	return config
}
