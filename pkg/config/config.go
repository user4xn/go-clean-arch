package config

import (
	"clean-arch/pkg/consts"
	"clean-arch/pkg/util"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		log.Println("No .env file found, using os environment variables")
	}

	viper.AutomaticEnv()
}

func GetRefreshDuration() int {
	jwtMode := util.GetEnv("JWT_MODE", "fallback")
	refreshDuration := consts.RefreshTokenDayAgeDev
	if jwtMode == "release" {
		refreshDuration = consts.RefreshTokenDayAgeRelease
	}
	return refreshDuration
}
