package di

import (
	"fmt"
	"github.com/mix-go/xcli"
	"github.com/mix-go/xdi"
	"github.com/mix-go/xutil/xenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func init() {
	obj := xdi.Object{
		Name: "logrus",
		New: func() (i interface{}, e error) {
			logger := logrus.New()
			logger.ReportCaller = true // 显示调用信息
			logFormat := xenv.Getenv("LOG_FORMAT").String("json")
			var filename string

			if logFormat == "json" {
				filename = fmt.Sprintf("%s/../runtime/logs/cli.json", xcli.App().BasePath)
				formatter := new(logrus.JSONFormatter)
				formatter.TimestampFormat = "2006-01-02 15:04:05.000"
				logger.Formatter = formatter
			} else {
				filename = fmt.Sprintf("%s/../runtime/logs/cli.log", xcli.App().BasePath)
				formatter := new(logrus.TextFormatter)
				formatter.TimestampFormat = "2006-01-02 15:04:05.000"
				formatter.FullTimestamp = true
				formatter.DisableQuote = true // 不转义换行符，为了保存错误堆栈到日志文件
				formatter.CallerPrettyfier = func(frame *runtime.Frame) (function string, file string) {
					return "", fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
				}
				logger.Formatter = formatter
			}

			fileRotate := &lumberjack.Logger{
				Filename:   filename,
				MaxBackups: 7,
			}
			writer := io.MultiWriter(os.Stdout, fileRotate)
			logger.SetOutput(writer)
			if xcli.App().Debug {
				logger.SetLevel(logrus.DebugLevel)
			}

			requestLogger := logger.WithFields(logrus.Fields{"_log": "cli"})
			return requestLogger, nil
		},
	}
	if err := xdi.Provide(&obj); err != nil {
		panic(err)
	}
}

func Logrus() (logger *logrus.Entry) {
	if err := xdi.Populate("logrus", &logger); err != nil {
		panic(err)
	}
	return
}
