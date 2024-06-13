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

func TestMySQL(t *testing.T) {
	config.LoadEnv("../../.env")

	assert.NotEmpty(t, config.MysqlHost())
	assert.NotEmpty(t, config.MysqlPort())
	assert.NotEmpty(t, config.MysqlUser())
	assert.NotEmpty(t, config.MysqlPass())
	assert.NotEmpty(t, config.MysqlDBName())
}
