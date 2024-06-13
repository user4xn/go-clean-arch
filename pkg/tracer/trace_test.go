package tracer_test

import (
	"clean-arch/pkg/tracer"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	logger, err := tracer.InitLogger("development")
	assert.Nil(t, err)
	assert.NotNil(t, logger)

	logger.Debug("Debug", zap.String("key", "value"))
	logger.Info("Info", zap.String("key", "value"))
	logger.Warn("Warn", zap.String("key", "value"))
	logger.Error("Error", zap.String("key", "value"))
	logger.Fatal("Fatal", zap.String("key", "value"))
}

func TestLoggingMiddleware(t *testing.T) {
	logger, err := tracer.InitLogger("development")
	assert.Nil(t, err)
	assert.NotNil(t, logger)

	middleware := tracer.LoggingMiddleware(logger)
	assert.NotNil(t, middleware)
}

func TestRecoverMiddleware(t *testing.T) {
	logger, err := tracer.InitLogger("development")
	assert.Nil(t, err)
	assert.NotNil(t, logger)

	middleware := tracer.RecoverMiddleware(logger)
	assert.NotNil(t, middleware)
}
