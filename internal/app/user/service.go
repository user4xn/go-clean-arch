package user

import (
	"clean-arch/database"
	"clean-arch/internal/dto"
	"clean-arch/internal/factory"
	"clean-arch/internal/model"
	"clean-arch/internal/repository"
	"clean-arch/pkg/consts"
	"clean-arch/pkg/util"
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type service struct {
	UserRepository repository.User
}

type Service interface {
	Store(ctx context.Context, reqHandler dto.PayloadUser) error
	FindAll(ctx context.Context, reqHandler dto.PayloadBasicTable) (*dto.ResponseUser, error)
	FindOne(ctx context.Context, id int) (dto.User, error)
	Update(ctx context.Context, id int, reqHandler dto.PayloadUpdateUser) error
	Delete(ctx context.Context, id int) error
}

func NewService(f *factory.Factory) Service {
	return &service{
		UserRepository: f.UserRepository,
	}
}

func (s *service) Store(ctx context.Context, reqHandler dto.PayloadUser) error {
	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)

	now := time.Now()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqHandler.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	insertModel := model.User{
		Name:            reqHandler.Name,
		Email:           reqHandler.Email,
		EmailVerifiedAt: &now,
		Password:        string(hashedPassword),
		PhoneNumber:     reqHandler.PhoneNumber,
	}

	if reqHandler.File != nil {
		insertModel.ProfileImageURL = reqHandler.URL
	}

	existingEmail, err := s.UserRepository.FindOne(ctx, "email", "email = ?", reqHandler.Email)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return err
		}
	}

	if existingEmail.Email != "" {
		return fmt.Errorf("email already exists")
	}

	if err := s.UserRepository.Store(tx, insertModel); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (s *service) FindAll(ctx context.Context, reqHandler dto.PayloadBasicTable) (*dto.ResponseUser, error) {
	var (
		total dto.ResponseTotalRow
		res   *dto.ResponseUser
		users []dto.User
		query string
		args  []interface{}
	)

	if reqHandler.Search != "" {
		query = "name LIKE ?"
		args = append(args, "%"+reqHandler.Search+"%")
	}

	count, err := s.UserRepository.Count(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	fetch, err := s.UserRepository.FindAll(ctx, "id, name, email, profile_image_url, email_verified_at, phone_number, created_at, updated_at", reqHandler.Limit, reqHandler.Offset, query, args...)
	if err != nil {
		return nil, err
	}

	for _, user := range fetch {
		var emailVerifiedAt *string
		if user.EmailVerifiedAt != nil {
			formatted := user.EmailVerifiedAt.Format(consts.TimeFormatDateTime)
			emailVerifiedAt = &formatted
		}

		users = append(users, dto.User{
			ID:              user.ID,
			Name:            user.Name,
			Email:           user.Email,
			EmailVerifiedAt: emailVerifiedAt,
			ProfileImageURL: user.ProfileImageURL,
			PhoneNumber:     user.PhoneNumber,
			CreatedAt:       user.CreatedAt.Format(consts.TimeFormatDateTime),
			UpdatedAt:       user.UpdatedAt.Format(consts.TimeFormatDateTime),
		})
	}

	total = dto.ResponseTotalRow{
		TotalRow: count,
	}

	res = &dto.ResponseUser{
		ResponseTotalRow: total,
		Data:             users,
	}

	return res, nil
}

func (s *service) Update(ctx context.Context, id int, reqHandler dto.PayloadUpdateUser) error {
	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)

	user, err := s.UserRepository.FindOne(ctx, "*", "id = ?", id)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	oldLink := user.ProfileImageURL
	appUrl := util.GetEnv("APP_URL", "http://localhost")
	appPort := util.GetEnv("APP_PORT", "8080")

	baseURL := fmt.Sprintf("%s:%s/", appUrl, appPort)
	sanitizedLink := strings.Replace(user.ProfileImageURL, baseURL, "", 1)

	if sanitizedLink != oldLink {
		err = util.DeleteFile(sanitizedLink)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	updatedModel := model.User{
		Name:        reqHandler.Name,
		PhoneNumber: reqHandler.PhoneNumber,
	}

	if reqHandler.File != nil {
		updatedModel.ProfileImageURL = reqHandler.URL
	}

	if reqHandler.NewPassword != "" {
		if reqHandler.LastPassword == "" {
			return fmt.Errorf("last password is required")
		}

		user, err := s.UserRepository.FindOne(ctx, "password", "id = ?", id)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("user not found")
			}

			return err
		}

		err = util.ComparePasswords(user.Password, reqHandler.LastPassword)
		if err != nil {
			return fmt.Errorf("last password does not match")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqHandler.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		updatedModel.Password = string(hashedPassword)
	}

	if err := s.UserRepository.UpdateOne(tx, id, updatedModel); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (s *service) FindOne(ctx context.Context, id int) (dto.User, error) {
	var (
		res dto.User
	)

	fetch, err := s.UserRepository.FindOne(ctx, "*", "id = ?", id)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return res, err
		}
	}

	var emailVerifiedAt *string
	if fetch.EmailVerifiedAt != nil { // Assuming EmailVerifiedAt is a *time.Time
		formatted := fetch.EmailVerifiedAt.Format(consts.TimeFormatDateTime)
		emailVerifiedAt = &formatted
	}

	res = dto.User{
		ID:              fetch.ID,
		Name:            fetch.Name,
		Email:           fetch.Email,
		EmailVerifiedAt: emailVerifiedAt,
		PhoneNumber:     fetch.PhoneNumber,
		ProfileImageURL: fetch.ProfileImageURL,
		CreatedAt:       fetch.CreatedAt.Format(consts.TimeFormatDateTime),
	}

	return res, nil
}

func (s *service) Delete(ctx context.Context, id int) error {
	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)

	user, err := s.UserRepository.FindOne(ctx, "*", "id = ?", id)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	oldLink := user.ProfileImageURL
	appUrl := util.GetEnv("APP_URL", "http://localhost")
	appPort := util.GetEnv("APP_PORT", "8080")

	baseURL := fmt.Sprintf("%s:%s/", appUrl, appPort)
	sanitizedLink := strings.Replace(user.ProfileImageURL, baseURL, "", 1)

	if err := s.UserRepository.DeleteOne(tx, id); err != nil {
		tx.Rollback()
		return err
	}

	if sanitizedLink != oldLink {
		err = util.DeleteFile(sanitizedLink)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	tx.Commit()
	return nil
}
