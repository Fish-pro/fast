package util

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func NewLogger() *logrus.Logger {
	logPath := "/tmp/fast-cni.log"
	if val, ok := os.LookupEnv("LOG_PATH"); ok {
		logPath = val
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	logger.SetFormatter(&logrus.TextFormatter{})
	return logger
}
