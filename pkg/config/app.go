package config

import "github.com/spf13/viper"

func AppEnv() string {
	return viper.GetString("APP_ENV")
}

func AppPort() int {
	return viper.GetInt("APP_PORT")
}

func AppSecretKey() string {
	return viper.GetString("APP_SECRET_KEY")
}

func AppApiKey() string {
	return viper.GetString("APP_API_KEY")
}
