package log

import (
	"app/di"
	"fmt"
)

func ErrHandle(err any) {
	//fmt.Println(string(debug.Stack()))
	di.Zap().Errorw(fmt.Sprintf("%v", err))
	di.Sentry().RecoverHandle(err)
}
