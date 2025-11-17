package auth

import (
	"bytes"
	"clean-arch/database"
	"clean-arch/internal/dto"
	"clean-arch/internal/factory"
	"clean-arch/internal/model"
	"clean-arch/internal/repository"
	"clean-arch/pkg/consts"
	"clean-arch/pkg/helper"
	"clean-arch/pkg/util"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"text/template"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

type service struct {
	UserRepository  repository.User
	OtpRepository   repository.Otp
	RedisRepository repository.Redis
	TwoFactor       bool
	TitleOTP        string
	TitleVerify     string
}

type Service interface {
	LoginAttempt(ctx context.Context, reqHandler dto.PayloadLoginTraced) (dto.ResponseJWT, error)
	VerifyEmail(ctx context.Context, base64String string) error
	RequestOTP(ctx context.Context, reqHandler dto.PayloadOtp) (dto.ResponseRequestOtp, error)
	VerifyOTP(ctx context.Context, reqHandler dto.PayloadVerifyOtpTraced) (any, error)
	Logout(ctx context.Context, bearer string) error
}

func NewService(f *factory.Factory) Service {
	return &service{
		TwoFactor:       true,
		UserRepository:  f.UserRepository,
		OtpRepository:   f.OtpRepository,
		RedisRepository: f.RedisRepository,
		TitleOTP:        "Kode Verifikasi " + util.GetEnv("APP_NAME", "fallback"),
		TitleVerify:     "Verifikasi Akun " + util.GetEnv("APP_NAME", "fallback"),
	}
}

func (s *service) Logout(ctx context.Context, bearer string) error {
	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return err
	}

	err := s.UserRepository.RevokeSession(tx, bearer)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (s *service) IncreaseAttempt(ctx context.Context, currentAttempt int, id int) error {
	updateOtp := model.OTP{
		Attempt: currentAttempt + 1,
	}

	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return err
	}

	err := s.OtpRepository.UpdateOne(tx, updateOtp, "id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (s *service) VerifyOTP(ctx context.Context, reqHandler dto.PayloadVerifyOtpTraced) (any, error) {
	var (
		res     dto.ResponseJWT
		resFail dto.ResponseFailVerifyOtp
	)

	now := time.Now()
	user, err := s.UserRepository.FindOne(ctx, "id, email, name, password, profile_image_url, email_verified_at", "email = ?", reqHandler.Email)
	if err != nil {
		return res, consts.UserNotFound
	}

	fetchOtp, err := s.OtpRepository.FindOne(ctx, true, "id, attempt, otp, expired_at", "user_id = ? AND expired_at > ?", user.ID, now)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("otp already expired, please request new one")
		}

		return nil, err
	}

	if fetchOtp.Attempt == 5 {
		return nil, fmt.Errorf("reached max verify attempt, please request a new one")
	}

	if fetchOtp.OTP != reqHandler.OTP {
		err = s.IncreaseAttempt(ctx, fetchOtp.Attempt, fetchOtp.ID)
		if err != nil {
			return res, fmt.Errorf("failed update attemps data otp %s", err.Error())
		}

		left := 5 - (fetchOtp.Attempt + 1)
		resFail = dto.ResponseFailVerifyOtp{
			AttemptLeft: fmt.Sprint(left),
		}

		return resFail, consts.OtpNotValid
	}

	secretKey := []byte(util.GetEnv("APP_SECRET_KEY", "fallback"))
	jwt, exp, err := s.GenerateToken(secretKey, strconv.Itoa(user.ID), user.Email)
	if err != nil {
		return res, consts.ErrorGenerateJwt
	}

	if jwt == "" {
		return res, consts.EmptyGenerateJwt
	}

	dataUser := dto.DataUserLogin{
		ID:              user.ID,
		Email:           user.Email,
		Name:            user.Name,
		EmailVerifiedAt: *user.EmailVerifiedAt,
		ProfileImageURL: user.ProfileImageURL,
	}

	sessionModel := model.UserSession{
		UserID:    user.ID,
		JWTToken:  jwt,
		ExpiresAt: *exp,
	}

	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return res, err
	}

	err = s.UserRepository.CreateSession(tx, sessionModel)
	if err != nil {
		tx.Rollback()
		return res, consts.ErrorGenerateJwt
	}
	tx.Commit()

	insertModel := model.LoginLog{
		UserID:    user.ID,
		IPAddress: reqHandler.IP,
		UserAgent: reqHandler.UserAgent,
	}

	tx = database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return res, err
	}

	err = s.UserRepository.StoreLoginLog(tx, insertModel)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error storing login logs %s", err.Error())
	}
	tx.Commit()

	cacheKey := fmt.Sprintf("user_session-%d", user.ID)
	_ = s.RedisRepository.Del(ctx, cacheKey)

	res = dto.ResponseJWT{
		TokenJwt:  jwt,
		ExpiredAt: exp.Format(consts.TimeFormatDateTime),
		DataUser:  &dataUser,
	}

	return res, nil
}

func (s *service) VerifyEmail(ctx context.Context, base64String string) error {
	emailDecode, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		fmt.Println("Error:", err)
		return consts.ErrorDecodeBase64
	}
	user, err := s.UserRepository.FindOne(ctx, "id, email_verified_at", "email = ?", emailDecode)
	if err != nil {
		return consts.NotFoundDataUser
	}

	if user.EmailVerifiedAt != nil {
		return nil
	}

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return consts.ErrorLoadLocationTime
	}

	now := time.Now().In(loc)

	updateUser := model.User{
		EmailVerifiedAt: &now,
	}

	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return err
	}

	err = s.UserRepository.UpdateOne(tx, user.ID, updateUser)
	if err != nil {
		tx.Rollback()
		log.Println("Error updating user:", err)
		return consts.FailedVerifyEmail
	}

	tx.Commit()

	return nil
}

func (s *service) GenerateToken(secretKey []byte, userID string, email string) (string, *time.Time, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return "", nil, err
	}

	jwtMode := util.GetEnv("JWT_MODE", "fallback")
	nonUnixTime := time.Now().In(loc).Add(consts.TokenDurationDev)
	expiredTime := nonUnixTime.Unix()

	if jwtMode == "release" {
		nonUnixTime := time.Now().In(loc).Add(consts.TokenDurationRelease)
		expiredTime := nonUnixTime.Unix()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"email":   email,
			"exp":     expiredTime,
		})

		tokenString, err := token.SignedString(secretKey)
		if err != nil {
			return "", nil, err
		}

		return tokenString, &nonUnixTime, nil
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     expiredTime,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &nonUnixTime, nil
}

func (s *service) SendVerifyEmail(user model.User) error {
	tmpl, err := template.ParseFiles(consts.TemplateEmailVerify)
	if err != nil {
		return fmt.Errorf("error parsing template %s", err.Error())
	}

	emailByte := []byte(user.Email)
	encodedString := base64.StdEncoding.EncodeToString(emailByte)
	urlVerify := "/auth/verify-email/"

	data := struct {
		AppUrl string
		Name   string
		Url    string
	}{
		AppUrl: util.GetEnv("APP_URL", "fallback") + ":" + util.GetEnv("APP_PORT", "fallback"),
		Name:   user.Name,
		Url:    util.GetEnv("FE_URL", "fallback") + urlVerify + encodedString,
	}

	var tplBuffer = new(bytes.Buffer)
	errExecute := tmpl.Execute(tplBuffer, data)
	if errExecute != nil {
		fmt.Println("Error executing template:", err)
		return consts.InvalidPassword
	}

	go helper.SendMail(user.Email, s.TitleVerify, tplBuffer.String())

	return nil
}

func (s *service) RequestOTP(ctx context.Context, reqHandler dto.PayloadOtp) (dto.ResponseRequestOtp, error) {
	var (
		res dto.ResponseRequestOtp
	)

	thisUser, err := s.UserRepository.FindOne(ctx, "id, email, name", "email = ?", reqHandler.Email)
	if err != nil {
		return res, consts.UserNotFound
	}

	now := time.Now()

	otp, err := util.GenerateOTP(6)
	if err != nil {
		return res, fmt.Errorf("error while generating OTP %x", err.Error())
	}

	countOtpToday, err := s.OtpRepository.CountOTP(ctx, "user_id = ? AND DATE(created_at) = ?", thisUser.ID, now.Format(consts.TimeFormatDate))
	if err != nil {
		return res, err
	}

	if countOtpToday == 0 {
		checkExpired, _ := s.OtpRepository.FindOne(ctx, true, "id, next_request_at, created_at", "user_id = ? AND expired_at < ?", thisUser.ID, now)
		if checkExpired.ID != 0 && now.Before(checkExpired.NextRequestAt) {
			res = dto.ResponseRequestOtp{
				LastRequestOn: checkExpired.CreatedAt.Format(consts.TimeFormatDateTime),
				NextRequestAt: checkExpired.NextRequestAt.Format(consts.TimeFormatDateTime),
			}

			return res, nil
		}
	}

	currentOtp, err := s.OtpRepository.FindOne(ctx, true, "id, attempt, expired_at, created_at, next_request_at", "user_id = ? AND expired_at > ?", thisUser.ID, now)
	if err != nil && err != gorm.ErrRecordNotFound {
		return res, err
	}

	if currentOtp.ID != 0 && currentOtp.NextRequestAt.After(now) {
		res = dto.ResponseRequestOtp{
			LastRequestOn: currentOtp.CreatedAt.Format(consts.TimeFormatDateTime),
			NextRequestAt: currentOtp.NextRequestAt.Format(consts.TimeFormatDateTime),
		}

		return res, nil
	}

	if countOtpToday == 5 {
		return res, consts.ErrorLimitOtp
	}

	expiredOtp := now.Add(time.Minute * 5)
	cooldown, _ := s.GetCooldownOtp(countOtpToday)
	nextRequest := now.Add(time.Second * time.Duration(cooldown))

	insertModel := model.OTP{
		UserID:        thisUser.ID,
		OTP:           otp,
		ExpiredAt:     expiredOtp,
		NextRequestAt: nextRequest,
	}

	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return res, err
	}

	err = s.OtpRepository.StoreOTP(tx, insertModel)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error while generating otp %s", err.Error())
	}
	tx.Commit()

	dataOtp := dto.DataOtpEmail{
		Name:  thisUser.Name,
		OTP:   otp,
		Email: thisUser.Email,
	}

	go s.SendOTPEmail(dataOtp)

	res = dto.ResponseRequestOtp{
		LastRequestOn: now.Format(consts.TimeFormatDateTime),
		NextRequestAt: nextRequest.Format(consts.TimeFormatDateTime),
	}

	return res, nil
}

func (s *service) GetCooldownOtp(countOtp int) (int, error) {
	switch countOtp {
	case 0:
		return int(consts.AttemptOne), nil
	case 1:
		return int(consts.AttemptTwo), nil
	case 2:
		return int(consts.AttemptThree), nil
	case 3:
		return int(consts.AttemptFour), nil
	case 4:
		return int(consts.AttemptFive), nil
	}

	return 0, fmt.Errorf("failed to get cooldown otp addon")
}

func (s *service) SendOTPEmail(dataOtp dto.DataOtpEmail) error {
	tmpl, err := template.ParseFiles(consts.TemplateEmailOtp)
	if err != nil {
		return fmt.Errorf("error parsing template %s", err.Error())
	}

	data := struct {
		AppUrl string
		Name   string
		Otp    string
	}{
		AppUrl: util.GetEnv("APP_URL", "fallback") + ":" + util.GetEnv("APP_PORT", "fallback"),
		Name:   dataOtp.Name,
		Otp:    dataOtp.OTP,
	}

	var tplBuffer = new(bytes.Buffer)
	errExecute := tmpl.Execute(tplBuffer, data)
	if errExecute != nil {
		fmt.Println("Error executing template:", err)
		return consts.InvalidPassword
	}

	go helper.SendMail(dataOtp.Email, s.TitleOTP, tplBuffer.String())

	return nil
}

func (s *service) Process2FA(ctx context.Context, body dto.PayloadLoginTraced, thisUser model.User) error {
	now := time.Now()

	loginLog, err := s.UserRepository.FindLoginLog(ctx, "ip_address = ? AND user_id = ? and DATE(created_at) = ?", body.IP, thisUser.ID, now.Format(consts.TimeFormatDate))
	if err != gorm.ErrRecordNotFound {
		return err
	}

	if loginLog.IPAddress == body.IP {
		return nil
	}

	return consts.Required2FA
}

func (s *service) LoginAttempt(ctx context.Context, reqHandler dto.PayloadLoginTraced) (dto.ResponseJWT, error) {
	var (
		res dto.ResponseJWT
	)

	user, err := s.UserRepository.FindOne(ctx, "id, email, name, profile_image_url, password, email_verified_at", "email = ?", reqHandler.Email)
	if err != nil {
		return res, consts.UserNotFound
	}

	err = util.ComparePasswords(user.Password, reqHandler.Password)
	if err != nil {
		return res, consts.InvalidPassword
	}

	if user.EmailVerifiedAt == nil {
		go s.SendVerifyEmail(user)

		return res, consts.UserNotVerifyEmail
	}

	if s.TwoFactor {
		err = s.Process2FA(ctx, reqHandler, user)
		if err != nil {
			return res, err
		}
	}

	secretKey := []byte(util.GetEnv("APP_SECRET_KEY", "fallback"))
	jwt, exp, err := s.GenerateToken(secretKey, strconv.Itoa(user.ID), user.Email)
	if err != nil {
		return res, consts.ErrorGenerateJwt
	}

	if jwt == "" {
		return res, consts.EmptyGenerateJwt
	}

	dataUser := dto.DataUserLogin{
		ID:              user.ID,
		Email:           user.Email,
		Name:            user.Name,
		EmailVerifiedAt: *user.EmailVerifiedAt,
		ProfileImageURL: user.ProfileImageURL,
	}

	sessionModel := model.UserSession{
		UserID:    user.ID,
		JWTToken:  jwt,
		ExpiresAt: *exp,
	}

	tx := database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return res, err
	}

	err = s.UserRepository.CreateSession(tx, sessionModel)
	if err != nil {
		tx.Rollback()
		return res, consts.ErrorGenerateJwt
	}
	tx.Commit()

	insertModel := model.LoginLog{
		UserID:    user.ID,
		IPAddress: reqHandler.IP,
		UserAgent: reqHandler.UserAgent,
	}

	tx = database.BeginTx(ctx, factory.NewFactory().InitDB)
	if err := tx.Error; err != nil {
		return res, err
	}

	err = s.UserRepository.StoreLoginLog(tx, insertModel)
	if err != nil {
		tx.Rollback()
		return res, fmt.Errorf("error storing login logs %s", err.Error())
	}
	tx.Commit()

	cacheKey := fmt.Sprintf("user_session-%d", user.ID)
	_ = s.RedisRepository.Del(ctx, cacheKey)

	res = dto.ResponseJWT{
		TokenJwt:  jwt,
		ExpiredAt: exp.Format(consts.TimeFormatDateTime),
		DataUser:  &dataUser,
	}

	return res, nil
}
