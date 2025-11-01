package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	APIVersion     string
	BaseAPIPath    string
	ServerAddress  string
	JWTSecret      string
	DBString       string
	StoragePath    string
	DeepseekAPIKey string
}

func LoadConfig() *Config {
	viper.AddConfigPath(".")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Err(err).Msg("Error reading env config file")
	} else {
		log.Info().Msg("Config file loaded")
	}

	apiVersion := viper.GetString("API_VERSION")
	BaseAPIPath := fmt.Sprintf("/api/%s", apiVersion)

	return &Config{
		APIVersion:     apiVersion,
		BaseAPIPath:    BaseAPIPath,
		ServerAddress:  viper.GetString("SERVER_ADDRESS"),
		JWTSecret:      viper.GetString("JWT_SECRET"),
		DBString:       viper.GetString("DB_STRING"),
		StoragePath:    viper.GetString("STORAGE_PATH"),
		DeepseekAPIKey: viper.GetString("DEEPSEEK_API_KEY"),
	}
}
