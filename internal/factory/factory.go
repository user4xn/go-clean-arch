package factory

import (
	"clean-arch/database"
	"clean-arch/internal/repository"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Factory struct {
	RedisClient     *redis.Client
	InitDB          *gorm.DB
	UserRepository  repository.User
	OtpRepository   repository.Otp
	RedisRepository repository.Redis
}

func NewFactory() *Factory {
	// Check db connection
	db := database.GetConnection()
	rdb := database.GetRedisClient()

	return &Factory{
		// Pass the db connection to repository package for database query calling
		RedisClient:     rdb,
		InitDB:          db,
		UserRepository:  repository.NewUserRepository(db),
		OtpRepository:   repository.NewOtpRepository(db),
		RedisRepository: repository.NewRedisRepository(rdb),
	}
}
