package di

import (
	"app/common/context"
	"fmt"
	"github.com/mix-go/xcli"
	"github.com/mix-go/xdi"
	"github.com/mix-go/xutil/xenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

func init() {
	obj := xdi.Object{
		Name: "zap",
		New: func() (i interface{}, e error) {
			atomicLevel := zap.NewAtomicLevelAt(zap.InfoLevel)
			logFormat := xenv.Getenv("LOG_FORMAT").String("json")
			var core zapcore.Core
			if logFormat == "json" {
				filename := fmt.Sprintf("%s/../runtime/logs/mix.json", xcli.App().BasePath)
				fileRotate := &lumberjack.Logger{
					Filename:   filename,
					MaxBackups: 7,
				}
				core = zapcore.NewCore(
					zapcore.NewJSONEncoder(zapcore.EncoderConfig{
						TimeKey:       "time",
						LevelKey:      "level",
						NameKey:       "logger",
						CallerKey:     "file",
						MessageKey:    "msg",
						StacktraceKey: "stackTrace",
						//LineEnding:  zapcore.DefaultLineEnding,
						EncodeLevel: zapcore.LowercaseLevelEncoder,
						EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
							enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
						},
						EncodeDuration: zapcore.StringDurationEncoder,
						EncodeCaller:   zapcore.ShortCallerEncoder,
					}),
					zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileRotate)),
					//zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileRotate)),
					atomicLevel,
				)
			} else {
				filename := fmt.Sprintf("%s/../runtime/logs/mix.log", xcli.App().BasePath)
				fileRotate := &lumberjack.Logger{
					Filename:   filename,
					MaxBackups: 7,
				}
				core = zapcore.NewCore(
					zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
						TimeKey:       "T",
						LevelKey:      "L",
						NameKey:       "N",
						CallerKey:     "C",
						MessageKey:    "M",
						StacktraceKey: "S",
						LineEnding:    zapcore.DefaultLineEnding,
						EncodeLevel:   zapcore.LowercaseLevelEncoder,
						EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
							enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
						},
						EncodeDuration: zapcore.StringDurationEncoder,
						EncodeCaller:   zapcore.ShortCallerEncoder,
					}),
					zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileRotate)),
					//zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileRotate)),
					atomicLevel,
				)
			}
			logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.WarnLevel))
			if xcli.App().Debug {
				atomicLevel.SetLevel(zap.DebugLevel)
			}
			return logger.Sugar().With(zap.String("_log", "mix")), nil
		},
	}
	if err := xdi.Provide(&obj); err != nil {
		panic(err)
	}
}

func Zap() (logger *zap.SugaredLogger) {
	if err := xdi.Populate("zap", &logger); err != nil {
		panic(err)
	}
	return
}

func ZapWithContext(ctx context.LogContextInterface) *zap.SugaredLogger {
	logArgs := ctx.GetLogArgs()
	if len(logArgs) > 0 {
		return Zap().With(logArgs...)
	}
	return Zap()
}

type ZapOutput struct {
	Logger *zap.SugaredLogger
}

func (t *ZapOutput) Write(p []byte) (n int, err error) {
	t.Logger.Debug(string(p))
	return len(p), nil
}
