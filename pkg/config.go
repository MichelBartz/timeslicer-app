package pkg

import (
	"fmt"

	"github.com/spf13/viper"
)

// TimeslicerConfig represents the configuration for the timeslicer-app
type TimeslicerConfig struct {
	Port               int
	StoreName          string
	TimeslicerInterval string
	TimeslicerStart    string
	TimeslicerEnd      string
}

// GetConfig returns the configuration for the timeslicer-app
func GetConfig() *TimeslicerConfig {
	viper.SetDefault("env", "dev")
	viper.SetDefault("port", 8080)

	configFileName := viper.GetString("env")
	viper.SetConfigName(configFileName)
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %s", err))
	}

	return &TimeslicerConfig{
		Port:               viper.GetInt("port"),
		StoreName:          viper.GetString("storeName"),
		TimeslicerInterval: viper.GetString("timeslice.interval"),
		TimeslicerStart:    viper.GetString("timeslice.start"),
		TimeslicerEnd:      viper.GetString("timeslice.end"),
	}
}
