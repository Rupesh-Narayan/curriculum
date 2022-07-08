package logger

import (
	"github.com/sirupsen/logrus"
	"log"
	"runtime"
)

// Client variable
var Client *logrus.Logger

func New() {
	Client = logrus.New()
	initialize()
}

func initialize() {
	formatter := new(logrus.TextFormatter)
	formatter.DisableColors = true
	formatter.FullTimestamp = true
	Client.SetFormatter(formatter)
	Client.SetReportCaller(true)
}

func GetErrorStack() string {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	log.Printf("Stack trace: %s", buf)
	return ""
}
