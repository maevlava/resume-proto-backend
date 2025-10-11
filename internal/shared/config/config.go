package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress string
	JWTSecret     string
	DBString      string
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

	return &Config{
		ServerAddress: viper.GetString("SERVER_ADDRESS"),
		JWTSecret:     viper.GetString("JWT_SECRET"),
		DBString:      viper.GetString("DB_STRING"),
	}
}
