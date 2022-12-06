package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ColectorServer string
	Facility       string
}

func Load() *Config {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var config Config
	config.ColectorServer = viper.GetString("colectorServer")
	config.Facility = viper.GetString("facility")
	return &config
}
