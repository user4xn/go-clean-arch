package tracer

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(env string) (*zap.Logger, error) {
	var cfg zap.Config

	switch strings.ToLower(env) {
	case "development":
		cfg = zap.Config{
			Encoding:         "console",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:     "message",
				LevelKey:       "level",
				TimeKey:        "time",
				CallerKey:      "caller",
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeLevel:    zapcore.CapitalColorLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}

	case "staging":
		cfg = zap.Config{
			Encoding:         "json",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:     "message",
				LevelKey:       "level",
				TimeKey:        "time",
				CallerKey:      "caller",
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}

	case "production":
		cfg = zap.Config{
			Encoding:         "json",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			EncoderConfig: zapcore.EncoderConfig{
				MessageKey:     "message",
				LevelKey:       "level",
				TimeKey:        "time",
				CallerKey:      "caller",
				EncodeDuration: zapcore.StringDurationEncoder,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			},
		}

	default:
		panic(fmt.Sprintf("Unknown app environment: %s", env))
	}

	return cfg.Build()
}

func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {

		logger = logger.With(zap.String("method", c.Request.Method))
		logger = logger.With(zap.String("url", c.Request.URL.String()))
		logger = logger.With(zap.String("host", c.Request.Host))
		logger = logger.With(zap.String("ip", c.ClientIP()))

		c.Set("log", logger)

		c.Next()
	}
}

func RecoverMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Fatal("Panic Recovered", zap.Any("error", err))
			}
		}()

		c.Next()
	}
}

func Log(c *gin.Context, key string, value interface{}) {
	logger, ok := c.Get("log")
	if !ok {
		return
	}

	logger.(*zap.Logger).With(zap.Any(key, value))

	c.Set("log", logger)
}
