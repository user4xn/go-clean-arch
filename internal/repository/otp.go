package repository

import (
	"clean-arch/internal/model"
	"clean-arch/pkg/util"
	"context"

	"gorm.io/gorm"
)

type Otp interface {
	StoreOTP(db *gorm.DB, insertModel model.OTP) error
	CountOTP(ctx context.Context, query string, args ...any) (int, error)
	FindOne(ctx context.Context, lastest bool, selectedFields string, query string, args ...any) (model.OTP, error)
	UpdateOne(db *gorm.DB, updateModel model.OTP, query string, args ...interface{}) error
}

type otp struct {
	Db *gorm.DB
}

func NewOtpRepository(db *gorm.DB) Otp {
	return &otp{
		Db: db,
	}
}

func (r *otp) UpdateOne(db *gorm.DB, updateModel model.OTP, query string, args ...interface{}) error {
	if err := db.Model(&model.OTP{}).Where(query, args...).Updates(updateModel).Error; err != nil {
		return err
	}

	return nil
}

func (r *otp) CountOTP(ctx context.Context, query string, args ...any) (int, error) {
	var (
		res int64
	)

	err := r.Db.WithContext(ctx).Model(model.OTP{}).Select("id").Where(query, args...).Count(&res).Error
	if err != nil {
		return 0, err
	}

	return int(res), nil
}

func (r *otp) StoreOTP(db *gorm.DB, insertModel model.OTP) error {
	if err := db.Model(model.OTP{}).Create(&insertModel).Error; err != nil {
		return err
	}

	return nil
}

func (r *otp) FindOne(ctx context.Context, lastest bool, selectedFields string, query string, args ...any) (model.OTP, error) {
	var res model.OTP

	db := r.Db.WithContext(ctx).Model(model.OTP{})
	db = util.SetSelectFields(db, selectedFields)

	if lastest {
		db = db.Order("created_at desc")
	}

	if err := db.Where(query, args...).Take(&res).Error; err != nil {
		return res, err
	}

	return res, nil
}
