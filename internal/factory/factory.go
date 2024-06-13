package factory

import (
	"clean-arch/database"
	"clean-arch/internal/repository"
)

type Factory struct {
	UserRepository repository.User
}

func NewFactory() *Factory {
	// Check db connection
	db := database.GetConnection()
	return &Factory{
		// Pass the db connection to repository package for database query calling
		repository.NewUserRepository(db),
	}
}