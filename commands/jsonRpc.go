package commands

import (
	"app/server/jsonRpc"
	"app/service"
	"github.com/mix-go/xcli/flag"
	"github.com/mix-go/xcli/process"
	"github.com/mix-go/xutil/xenv"
)

type JsonRpcCommand struct {
}

func (t *JsonRpcCommand) Main() {
	if flag.Match("d", "daemon").Bool() {
		process.Daemon()
	}

	addr := xenv.Getenv("JSON_RPC_ADDR").String(":8082")
	network := xenv.Getenv("JSON_RPC_NETWORK").String("tcp")

	server := jsonRpc.NewJsonRpcServer()
	server.Register("Hello", &service.HelloService{})

	welcome("mix-json-rpc", addr)
	server.Run(addr, network)
}
