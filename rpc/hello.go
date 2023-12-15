package rpc

import (
	"app/common/context"
	"app/common/jsonRpcHelper"
)

type HelloRpc struct {
}

func (t *HelloRpc) Index(ctx *context.RunContext, params map[string]any) (data any, err *jsonRpcHelper.Error) {
	data = "hello world"
	//err = jsonRpcHelper.NewError(-1111, "test")
	//time.Sleep(5 * time.Second)
	return
}
