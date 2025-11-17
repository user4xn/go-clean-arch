package repository

import (
	"clean-arch/internal/model"
	"clean-arch/pkg/consts"
	"clean-arch/pkg/util"
	"context"

	"gorm.io/gorm"
)

type User interface {
	FindAll(ctx context.Context, selectedFields string, limit, offset int, query string, args ...interface{}) ([]*model.User, error)
	FindOne(ctx context.Context, selectedFields string, query string, args ...any) (model.User, error)
	Store(db *gorm.DB, insertModel model.User) error
	UpdateOne(db *gorm.DB, id int, data model.User) error
	UpdateAll(db *gorm.DB, data model.User, selectedFields string, query string, args ...any) error
	DeleteOne(db *gorm.DB, id int) error
	Count(ctx context.Context, query string, args ...any) (int, error)

	FindSession(ctx context.Context, userId int, token string) (model.UserSession, error)
	CreateSession(db *gorm.DB, sessionData model.UserSession) error
	FindLoginLog(ctx context.Context, query string, args ...any) (model.LoginLog, error)
	StoreLoginLog(db *gorm.DB, insertModel model.LoginLog) error
	RevokeSession(db *gorm.DB, bearer string) error
}

type user struct {
	Db *gorm.DB
}

func NewUserRepository(db *gorm.DB) User {
	return &user{
		Db: db,
	}
}

func (r *user) RevokeSession(db *gorm.DB, bearer string) error {
	modelUpdate := model.UserSession{
		Revoked: 1,
	}

	err := db.Model(model.UserSession{}).Where("jwt_token = ?", bearer).Updates(modelUpdate).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *user) StoreLoginLog(db *gorm.DB, insertModel model.LoginLog) error {
	if err := db.Model(model.LoginLog{}).Create(&insertModel).Error; err != nil {
		return err
	}

	return nil
}

func (r *user) FindLoginLog(ctx context.Context, query string, args ...any) (model.LoginLog, error) {
	var (
		res model.LoginLog
	)

	err := r.Db.WithContext(ctx).Model(&model.LoginLog{}).Where(query, args...).Take(&res).Error
	if err != nil {
		return res, err
	}

	return res, nil
}

func (r *user) Store(db *gorm.DB, insertModel model.User) error {
	if err := db.Model(model.User{}).Create(&insertModel).Error; err != nil {
		return err
	}

	return nil
}

func (r *user) FindAll(ctx context.Context, selectedFields string, limit, offset int, query string, args ...interface{}) ([]*model.User, error) {
	var res []*model.User
	db := r.Db.Model(&model.User{})
	db = util.SetSelectFields(db, selectedFields)
	db = db.Where(query, args...).Debug()

	if limit > 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err := db.Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *user) FindOne(ctx context.Context, selectedFields string, query string, args ...any) (model.User, error) {
	var res model.User

	db := r.Db.WithContext(ctx).Model(model.User{})
	db = util.SetSelectFields(db, selectedFields)

	if err := db.Where(query, args...).Take(&res).Error; err != nil {
		return res, err
	}

	return res, nil
}

func (r *user) CreateSession(db *gorm.DB, sessionData model.UserSession) error {
	if err := db.Model(model.UserSession{}).Create(&sessionData).Error; err != nil {
		return err
	}

	return nil
}

func (r *user) FindSession(ctx context.Context, userId int, token string) (model.UserSession, error) {
	var res model.UserSession

	db := r.Db.WithContext(ctx).Model(model.UserSession{})
	if err := db.Where("user_id = ? AND jwt_token = ? AND revoked = ?", userId, token, consts.SessionActive).Take(&res).Error; err != nil {
		return model.UserSession{}, err
	}

	return res, nil
}

func (r *user) UpdateOne(db *gorm.DB, id int, data model.User) error {
	if err := db.Model(&model.User{}).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *user) UpdateAll(db *gorm.DB, data model.User, selectedFields string, query string, args ...any) error {
	if err := db.Model(&model.User{}).Select(selectedFields).Where(query, args).Debug().Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func (r *user) Count(ctx context.Context, query string, args ...any) (int, error) {
	var (
		res int64
	)

	err := r.Db.WithContext(ctx).Model(model.User{}).Select("id").Where(query, args...).Count(&res).Error
	if err != nil {
		return 0, err
	}

	return int(res), nil
}

func (r *user) DeleteOne(db *gorm.DB, id int) error {
	if err := db.Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}
