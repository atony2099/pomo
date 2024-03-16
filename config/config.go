package config

import (
	"log"

	"github.com/spf13/viper"
)

type Configuration struct {
	DBDSN        string
	RedisURL     string
	AuthKey      string
	PomodoroTime int
	StopInFirst  int
	BreakTime    int
	TeamID       string
}

func LoadConfig() *Configuration {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("$HOME/.config/pomo") //
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	conf := &Configuration{}
	if err := viper.Unmarshal(conf); err != nil {
		log.Fatalf("Failed to unmarshal configuration: %s", err)
	}

	return conf
}
