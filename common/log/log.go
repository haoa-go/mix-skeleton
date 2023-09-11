package log

import (
	"app/common/context"
	"app/di"
	"fmt"
)

func ErrHandle(err any) {
	//fmt.Println(string(debug.Stack()))
	di.Zap().Errorw(fmt.Sprintf("%v", err))
	di.Sentry().ErrHandle(err)
}

func ErrHandleWithContext(ctx context.LogContextInterface, err any) {
	//fmt.Println(string(debug.Stack()))
	di.ZapWithContext(ctx).Errorw(fmt.Sprintf("%v", err))
	di.Sentry().ErrHandle(err)
}

func RecoverHandle() {
	if err := recover(); err != nil {
		ErrHandle(err)
	}
}
