package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

func Reader() error {
	// Set config path
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/NBCFB/config/")
	viper.SetConfigType("json")

	// Check if file has been changed
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file changed:", e.Name)
	})

	// Read config file
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}