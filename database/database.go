package database

import (
	"clean-arch/pkg/config"
	"clean-arch/pkg/util"
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	dbConn  *gorm.DB
	rdb     *redis.Client
	once    sync.Once
	onceRdb sync.Once
)

func CreateConnection() {
	once.Do(func() {
		dbType := util.GetEnv("DB_DRIVER", "mysql")

		conf := dbConfig{
			User: config.DbUser(),
			Pass: config.DbPass(),
			Host: config.DbHost(),
			Port: config.DbPort(),
			Name: config.DbName(),
		}

		switch dbType {
		case "psql":
			pg := postgresConfig{
				dbConfig: conf,
				SSLMode:  util.GetEnv("DB_SSLMODE", "disable"),
			}
			pg.Connect()

		default:
			mysql := mysqlConfig{dbConfig: conf}
			mysql.Connect()
		}
	})
}

func GetConnection() *gorm.DB {
	if dbConn == nil {
		CreateConnection()
	}
	return dbConn
}

func BeginTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx).Begin()
}

func GetRedisClient() *redis.Client {

	onceRdb.Do(func() {
		rdb = redis.NewClient(&redis.Options{
			Addr:     util.GetEnv("REDIS_HOST", "localhost:6379"),
			Password: util.GetEnv("REDIS_PASSWORD", ""),
			DB:       0,
		})
	})

	return rdb
}
