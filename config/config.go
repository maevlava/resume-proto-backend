package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress string
}

func LoadConfig() *Config {
	viper.AddConfigPath("./config")
	viper.SetConfigName("dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Error reading config file")
		return nil
	}

	return &Config{
		ServerAddress: viper.GetString("SERVER_ADDRESS"),
	}
}
