package config

import "github.com/spf13/viper"

func DbHost() string {
	return viper.GetString("DB_HOST")
}

func DbPort() string {
	return viper.GetString("DB_PORT")
}

func DbUser() string {
	return viper.GetString("DB_USER")
}

func DbPass() string {
	return viper.GetString("DB_PASS")
}

func DbName() string {
	return viper.GetString("DB_NAME")
}
