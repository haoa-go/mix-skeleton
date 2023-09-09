package di

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/mix-go/xcli"
	"github.com/mix-go/xdi"
	"github.com/spf13/viper"
	"time"
)

func init() {
	obj := xdi.Object{
		Name: "sentry",
		New: func() (i interface{}, e error) {
			dsn := viper.GetString("sentry-dsn")
			if dsn == "" {
				return &SentryHelper{}, nil
			}
			sentryTransport := sentry.NewHTTPTransport()
			sentryTransport.Timeout = 3 * time.Second
			err := sentry.Init(sentry.ClientOptions{
				Dsn:              dsn,
				Debug:            xcli.App().Debug,
				Transport:        sentryTransport,
				AttachStacktrace: true,
			})
			if err != nil {
				Zap().Errorf("sentry.Init: %v", err)
			}
			return &SentryHelper{}, nil
		},
	}
	if err := xdi.Provide(&obj); err != nil {
		panic(err)
	}
}

func Sentry() (s *SentryHelper) {
	if err := xdi.Populate("sentry", &s); err != nil {
		panic(err)
	}
	return
}

type SentryHelper struct {
}

func (t *SentryHelper) RecoverHandle(err any) {
	if viper.GetString("sentry-dsn") == "" {
		return
	}
	localHub := sentry.CurrentHub().Clone()
	defer func() {
		if err := recover(); err != nil {
			Zap().Error(fmt.Sprintf("%v", err))
		}
	}()
	localHub.Recover(err)
	localHub.Flush(time.Second * 2)
}

func (t *SentryHelper) Recover() {
	if err := recover(); err != nil {
		t.RecoverHandle(err)
	}
}
