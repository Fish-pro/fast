package util

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func NewLogger() *logrus.Logger {
	logPath := "/var/log/fast/cni/cni.log"
	if val, ok := os.LookupEnv("LOG_PATH"); ok {
		logPath = val
	}

	file, err := os.OpenFile(logPath, os.O_APPEND, 0666)
	if err != nil {
		logger.Info("failed to open log file")
	}

	logger.SetFormatter(&logrus.TextFormatter{})
	logger.Out = file
	logrus.Info("init logger successfully")
	return logger
}
