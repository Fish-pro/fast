package util

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger()
	logger.WithFields(logrus.Fields{
		"key":  "value",
		"key1": "value1",
	}).Info("info")
	logger.WithError(fmt.Errorf("this is a error")).Error("err")
}

func BenchmarkNeLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		logger := NewLogger()
		logger.WithFields(logrus.Fields{
			"i":    i,
			"key":  "value",
			"key1": "value1",
		}).Info("info")
		logger.WithError(fmt.Errorf("this is a error")).Error("err")
	}
}
