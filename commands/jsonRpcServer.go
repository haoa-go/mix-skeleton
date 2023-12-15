package commands

import (
	"app/rpc"
	"app/server/jsonRpc"
	"github.com/mix-go/xcli/flag"
	"github.com/mix-go/xcli/process"
	"github.com/mix-go/xutil/xenv"
)

type JsonRpcServerCommand struct {
}

func (t *JsonRpcServerCommand) Main() {
	if flag.Match("d", "daemon").Bool() {
		process.Daemon()
	}

	addr := xenv.Getenv("JSON_RPC_ADDR").String(":8082")
	network := xenv.Getenv("JSON_RPC_NETWORK").String("tcp")

	server := jsonRpc.NewJsonRpcServer()
	server.Register("Hello", &rpc.HelloRpc{})

	welcome("mix-json-rpc", addr)
	server.Run(addr, network)
}
