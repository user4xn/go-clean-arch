package config

import "github.com/spf13/viper"

func MysqlHost() string {
	return viper.GetString("MYSQL_HOST")
}

func MysqlPort() string {
	return viper.GetString("MYSQL_PORT")
}

func MysqlUser() string {
	return viper.GetString("MYSQL_USER")
}

func MysqlPass() string {
	return viper.GetString("MYSQL_PASS")
}

func MysqlDBName() string {
	return viper.GetString("MYSQL_DB_NAME")
}
