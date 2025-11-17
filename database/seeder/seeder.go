package seeder

import (
	"clean-arch/database"
	"clean-arch/internal/model"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed() {
	var (
		err error
	)

	db := database.GetConnection()

	err = UserSeed(db)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func UserSeed(db *gorm.DB) error {
	var (
		InsertModel []model.User
	)

	fmt.Println("executing user seed...")

	now := time.Now()
	passwordAdmin := []byte("demouser123")
	hashedPasswordAdmin, err := bcrypt.GenerateFromPassword(passwordAdmin, bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		return err
	}

	InsertModel = []model.User{
		{
			Name:            "demouser",
			Email:           "demouser@gmail.com",
			EmailVerifiedAt: &now,
			Password:        string(hashedPasswordAdmin),
			PhoneNumber:     "08123456789",
		},
	}

	for _, data := range InsertModel {
		model := model.User{}
		err := db.Where("email = ?", data.Email).First(&model).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			if err := db.Create(&data).Error; err != nil {
				fmt.Println(err)
				return err
			}
		}
	}

	fmt.Print("success executing user seed...")

	return nil
}
