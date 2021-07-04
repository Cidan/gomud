package config

import "github.com/spf13/viper"

func init() {
	viper.AddConfigPath(".")
	viper.SetDefault("save_path", "/tmp/")
}

// Set will set a config value for the given key.
func Set(key string, value interface{}) {
	viper.Set(key, value)
}

// GetString gets a config key's value as a string.
func GetString(key string) string {
	return viper.GetString(key)
}
