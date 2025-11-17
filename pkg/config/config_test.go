package config_test

import (
	"clean-arch/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	config.LoadEnv("../../.env")

	assert.NotEmpty(t, config.AppEnv())
	assert.NotEmpty(t, config.AppPort())
	assert.NotEmpty(t, config.AppSecretKey())
}

func TestMyDB(t *testing.T) {
	config.LoadEnv("../../.env")

	assert.NotEmpty(t, config.DbHost())
	assert.NotEmpty(t, config.DbPort())
	assert.NotEmpty(t, config.DbUser())
	assert.NotEmpty(t, config.DbPass())
	assert.NotEmpty(t, config.DbName())
}
