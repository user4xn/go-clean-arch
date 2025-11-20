package middleware

import (
	"clean-arch/internal/dto"
	"clean-arch/internal/factory"
	"clean-arch/pkg/dbutil"
	"clean-arch/pkg/util"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header["Authorization"]

		if len(header) == 0 {
			response := util.APIResponse("Sorry, you didn't enter a valid bearer token", http.StatusUnauthorized, "failed", nil)
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		rep := regexp.MustCompile(`(Bearer)\s?`)
		bearerStr := rep.ReplaceAllString(header[0], "")
		parsedToken, err := parseToken(bearerStr)
		if err != nil || !parsedToken.Valid {
			response := util.APIResponse("Unauthorized, bearer token not valid", http.StatusUnauthorized, "failed", nil)
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		claims := parsedToken.Claims.(jwt.MapClaims)

		f := factory.NewFactory()
		userId, _ := strconv.Atoi(claims["user_id"].(string))

		_, err = f.UserRepository.FindSession(c, bearerStr)
		if err != nil {
			response := util.APIResponse("Unauthorized", http.StatusUnauthorized, "failed", nil)
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		cacheKey := fmt.Sprintf("user_session-%d", userId)

		cachedData, err := f.RedisClient.Get(c, cacheKey).Result()
		if err == nil {
			var cachedInfo dto.JwtSession
			if err := json.Unmarshal([]byte(cachedData), &cachedInfo); err == nil {
				c.Set("user", cachedInfo)
			}
		}

		if err == redis.Nil {
			var jwtSess dto.JwtSession
			user, _ := f.UserRepository.FindOne(c, "*", dbutil.Where("id = ?", userId))

			jwtSess = dto.JwtSession{
				ID:          user.ID,
				Name:        user.Name,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
				CreatedAt:   user.CreatedAt,
			}

			jsonData, err := json.Marshal(jwtSess)
			if err == nil {
				f.RedisClient.Set(c, cacheKey, jsonData, time.Hour)
			} else {
				fmt.Println("Error marshalling data for cache:", err)
			}

			c.Set("user", jwtSess)
		}

		c.Set("bearer", bearerStr)

		c.Next()
	}
}

func parseToken(tokenString string) (*jwt.Token, error) {
	secretKey := []byte(util.GetEnv("APP_SECRET_KEY", "fallback"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secretKey, nil
	})

	return token, err
}
