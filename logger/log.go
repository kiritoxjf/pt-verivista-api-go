package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

// SetLogOutPut 设置日志输出
func SetLogOutPut() error {
	// 打开或创建日志文件
	logFilePath := os.Getenv("VERIVISTA_LOG_PATH")
	if logFilePath == "" {
		logFilePath = "logger/verivista.log"
	}
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("[Open Log File Error]: %v", err)
	}
	// 设置日志输出文件
	mw := io.MultiWriter(os.Stdout, file)
	logrus.SetOutput(mw)

	return nil
}
