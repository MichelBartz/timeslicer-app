package pkg

import (
	"fmt"

	"github.com/spf13/viper"
)

type TimeslicerConfig struct {
	Port               int
	StoreName          string
	TimeslicerInterval string
	TimeslicerStart    string
	TimeslicerEnd      string
}

func GetConfig() *TimeslicerConfig {
	viper.SetDefault("env", "dev")
	viper.SetDefault("port", 8080)

	configFileName := viper.GetString("env")
	viper.SetConfigName(configFileName)
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	return &TimeslicerConfig{
		Port:               viper.GetInt("port"),
		StoreName:          viper.GetString("storeName"),
		TimeslicerInterval: viper.GetString("timeslice.interval"),
		TimeslicerStart:    viper.GetString("timeslice.start"),
		TimeslicerEnd:      viper.GetString("timeslice.end"),
	}
}
