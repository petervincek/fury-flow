package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/petervincek/fury-flow/internal/logging"
	"go.uber.org/zap"
)

var logger = logging.GetLogger()

func TryLoadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		// Check if the error is because the .env file does not exist
		if os.IsNotExist(err) {
			logger.Warn("No .env file found, skipping loading environment variables.")
		} else {
			logger.Fatal("Error while loading default env file", zap.Error(err))
		}
	}
}
