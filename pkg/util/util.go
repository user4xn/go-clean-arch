package util

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func GetEnv(key string, fallback string) string {
	// Godotenv read the .env file on the root folder
	a, _ := godotenv.Read()
	var (
		val     string
		isExist bool
	)
	// Check the key of the env using Hashmap
	// if exist return the actual value, if !exist return the fallback value
	val, isExist = a[key]
	if !isExist {
		val = fallback
	}
	return val
}

type Response struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data"`
}

type Meta struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

func FormatValidationError(err error) []string {
	var dataErrror []string
	var foo *json.UnmarshalTypeError
	if errors.As(err, &foo) {
		dataErrror = append(dataErrror, err.Error())
		return dataErrror
	}
	for _, e := range err.(validator.ValidationErrors) {
		dataErrror = append(dataErrror, e.Error())
	}

	return dataErrror
}

func APIResponse(message string, code int, status string, data interface{}) Response {
	meta := Meta{
		Message: message,
		Code:    code,
		Status:  status,
	}

	jsonResponse := Response{
		Meta: meta,
		Data: data,
	}

	return jsonResponse

}

func GenerateRefreshToken() (string, error) {
	refreshToken := make([]byte, 32)
	_, err := rand.Read(refreshToken)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", refreshToken), nil
}

func CreateErrorLog(errMessage error) {
	fileName := fmt.Sprintf("./storage/error_logs/error-%s.log", time.Now().Format("2006-01-02"))

	// open log file
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	defer logFile.Close()

	// set log out put
	log.SetOutput(logFile)

	log.SetFlags(log.LstdFlags)

	_, fileName, line, _ := runtime.Caller(1)
	log.Printf("[Error] in [%s:%d] %v", fileName, line, errMessage.Error())
}

func GenerateOTP(length int) (string, error) {
	otpChars := "0123456789"
	otp := make([]byte, length)

	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(otpChars))))
		if err != nil {
			return "", err
		}
		otp[i] = otpChars[num.Int64()]
	}
	return string(otp), nil
}

func IntSliceContains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func InArrayStr(array []string, str string) bool {
	for _, v := range array {
		if v == str {
			return true
		}
	}

	return false
}

func SaveFile(file *multipart.FileHeader) (string, error) {
	uniqueName := fmt.Sprintf("%d-%s", time.Now().Unix(), file.Filename)
	savePath := filepath.Join("uploads", uniqueName)

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	appUrl := GetEnv("APP_URL", "http://localhost")
	appPort := GetEnv("APP_PORT", "8080")

	fileURL := fmt.Sprintf("%s:%s/%s", appUrl, appPort, savePath)
	return fileURL, nil
}

func DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return err // Return error if file deletion fails
	}
	return nil
}

func StripHTML(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return strings.TrimSpace(re.ReplaceAllString(input, ""))
}

/*********/

func BuildLikeClause(columns []string, keyword string) (string, []interface{}) {
	likeKeyword := fmt.Sprintf("%%%s%%", strings.ToLower(keyword))

	var conditions []string
	var params []interface{}

	for _, column := range columns {
		conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE ?", column))
		params = append(params, likeKeyword)
	}

	return strings.Join(conditions, " OR "), params
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
