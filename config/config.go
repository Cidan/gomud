package config

import (
	"sync"

	"github.com/spf13/viper"
)

var mutex sync.RWMutex

func init() {
	viper.AddConfigPath(".")
	viper.SetDefault("save_path", "/tmp")
	mutex = sync.RWMutex{}
}

// Set will set a config value for the given key.
func Set(key string, value interface{}) {
	mutex.Lock()
	defer mutex.Unlock()
	viper.Set(key, value)
}

// GetString gets a config key's value as a string.
func GetString(key string) string {
	mutex.RLock()
	defer mutex.RUnlock()
	return viper.GetString(key)
}
