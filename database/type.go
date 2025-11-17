package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Define the database conn configuration
type (
	dbConfig struct {
		Host string
		User string
		Pass string
		Port string
		Name string
	}

	mysqlConfig struct {
		dbConfig
	}

	postgresConfig struct {
		dbConfig
		SSLMode string
	}
)

var err error

// Connect to mysql with the input configuration
func (conf mysqlConfig) Connect() {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s",
		conf.User,
		conf.Pass,
		conf.Host,
		conf.Port,
		conf.Name,
		"utf8mb4",
		"True",
		"Local",
	)

	dbConn, err = gorm.Open(mysql.New(mysql.Config{
		DriverName:           "mysql",
		DisableWithReturning: true,
		DSN:                  dsn,
	}), &gorm.Config{
		SkipDefaultTransaction:   true,
		DisableNestedTransaction: true,
	})
	if err != nil {
		panic(err)
	}
}

func (p *postgresConfig) Connect() {
	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		p.User, p.Pass,
		p.Host, p.Port,
		p.Name, p.SSLMode,
	)

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	dbConn = conn
}
