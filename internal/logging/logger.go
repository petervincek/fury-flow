package logging

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/petervincek/fury-flow/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger     *zap.Logger
	onceLogger sync.Once
)

// GetLogger returns a singleton zap.Logger instance
func GetLogger() *zap.Logger {
	onceLogger.Do(func() {
		config.LoadEnvVariables(".env")

		zapCores := []zapcore.Core{}
		fileZapCore, ok := tryToCreateFileZapCore()
		fmt.Printf("FileZapCore created: %t\n", ok)
		if ok {
			zapCores = append(zapCores, fileZapCore)
		}
		zapCores = append(zapCores, createConsoleZapCore())
		core := zapcore.NewTee(zapCores...)
		// Create the logger
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	})
	return logger
}

// getLogLevel retrieves the log level from the "LOG_LEVEL" environment variable.
// It returns the corresponding zapcore.Level and an error if the value is invalid.
// Supported log levels are: "debug", "info", "warn", "error", and "fatal".
// If the environment variable is not set or contains an invalid value, it defaults to zapcore.DebugLevel and returns an error.
func getLogLevel() (zapcore.Level, error) {
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.DebugLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// tryToCreateFileZapCore attempts to create a zapcore.Core for logging to a file.
// It reads the log file path from the "LOG_FILE" environment variable. If the variable
// is not set or the file cannot be opened, it returns (nil, false).
// The log format is determined by the "LOG_FORMAT" environment variable: if set to "text",
// logs are written in a human-readable format; otherwise, logs are written in JSON format.
// The log level is obtained via getLogLevel(). On success, returns the configured zapcore.Core
// and true; otherwise, returns nil and false.
func tryToCreateFileZapCore() (zapcore.Core, bool) {
	logFilePath := os.Getenv("LOG_FILE")
	if logFilePath == "" {
		// in this case there is no support for logging into the file
		return nil, false
	}
	// Create a file writer
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, false
	}

	// Create a zapcore that writes to the file
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "text" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig) // Human-readable format
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig) // JSON format
	}

	logLevel, err := getLogLevel()
	if err != nil {
		return nil, false
	}
	fileCore := zapcore.NewCore(
		encoder,               // Encoder based on LOG_FORMAT
		zapcore.AddSync(file), // Write to file
		logLevel,              // Log level
	)
	return fileCore, true
}

// createConsoleZapCore initializes and returns a zapcore.Core configured for console output.
// It sets up a human-readable encoder, writes logs to standard error, and applies the current log level.
// This core is suitable for development environments where readable log output is preferred.
func createConsoleZapCore() zapcore.Core {
	logLevel, _ := getLogLevel()
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), // Human-readable format
		zapcore.AddSync(os.Stderr),                                   // Write to stderr
		logLevel,                                                     // Log level
	)
	return consoleCore
}
