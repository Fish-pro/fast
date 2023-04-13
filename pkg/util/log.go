package util

import (
	"bufio"
	"os"
)

var logPath string

func init() {
	val, ok := os.LookupEnv("LOG_PATH")
	if ok {
		logPath = val
	} else {
		logPath = "/var/log/fast/cni/cni.log"
	}
}

func WriteLog(log ...string) {
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		os.Create(logPath)
	}
	defer file.Close()

	write := bufio.NewWriter(file)
	logRes := ""
	for _, c := range log {
		logRes += c
		logRes += " "
	}

	_, err = write.WriteString(logRes + "\r\n")
	if err != nil {
		return
	}

	write.Flush()
}
