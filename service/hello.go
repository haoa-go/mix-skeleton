package service

import (
	"app/common/context"
	"app/common/jsonRpcHelper"
	"app/exception"
	"fmt"
)

type HelloService struct {
}

func (t *HelloService) Index(ctx *context.RunContext, params map[string]any) (data any, err *jsonRpcHelper.Error) {
	data = "hello world"
	fmt.Printf("params: %v\n", params)
	//err = jsonRpcHelper.NewError(-1111, "test")
	panic(exception.NewMsgEmptyData("test err", -111))
	return
}
