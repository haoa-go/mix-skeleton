package jsonRpc

import (
	context2 "app/common/context"
	"app/common/jsonRpcHelper"
	"app/common/log"
	"app/di"
	"app/exception"
	"app/service"
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
)

const (
	RPC_ID_FIELD = "rpc_id"
)

type JsonRpcServer struct {
	serviceMap     map[string]any
	structValueMap map[string]reflect.Value
	callValueMap   map[string]reflect.Value
}

func NewJsonRpcServer() *JsonRpcServer {
	return &JsonRpcServer{
		serviceMap:     make(map[string]any),
		structValueMap: make(map[string]reflect.Value),
		callValueMap:   make(map[string]reflect.Value),
	}
}

func (t *JsonRpcServer) Run(addr string, network string) {
	if network == "unix" {
		removeErr := os.Remove(addr)
		if removeErr != nil && !strings.Contains(removeErr.Error(), "no such file or directory") {
			di.Zap().Error(removeErr)
			return
		}
	}

	listen, err := net.Listen(network, addr)
	if err != nil {
		di.Zap().Errorf("listen error: %v", err)
		return
	}

	// signal
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-ch
		di.Zap().Info("Server shutdown")
		if err := listen.Close(); err != nil {
			di.Zap().Errorf("Server shutdown error: %s", err)
		}
	}()

	for {
		// 接收客户端向服务端建立的连接
		conn, err := listen.Accept() // 可以与客户端段建立连接，如果没有连接就挂起阻塞

		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			di.Zap().Warnf("accept err: %v", err)
			return
		}
		//处理客户连接
		go t.handler(conn) // 可以利用协程处理，提高效率
	}
}

func (t *JsonRpcServer) Register(name string, obj any) {
	t.serviceMap[name] = obj
}

func (t *JsonRpcServer) handler(conn net.Conn) {
	ctx := context2.NewRunContext()
	defer func() {
		ctx.Clear()
	}()
	defer t.recover(ctx, conn)
	reader := bufio.NewReader(conn)

	//for {
	//	unpack2(conn, reader)
	//}

	for {
		// 接收
		received, err := t.unpack(reader)
		//di.Zap().Debugf("received: %s", string(received))
		if err != nil {
			if err != io.EOF {
				di.Zap().Errorf("unpack error: %v", err)
			}
			conn.Close()
			break
		}

		id, rpcError, resData := t.dispatch(ctx, received)
		if rpcError != nil {
			t.sendResponse(conn, t.buildErrorJson(rpcError, id))
		} else {
			t.sendResponse(conn, t.buildSuccessJson(resData, id))
		}
	}
}

func (t *JsonRpcServer) unpack(reader *bufio.Reader) ([]byte, error) {
	// 读取数据长度
	lenBuf := make([]byte, 4)
	_, err := reader.Read(lenBuf)
	if err != nil {
		return []byte(""), err
	}
	packetLen := binary.BigEndian.Uint32(lenBuf)

	// 根据长度读取数据内容
	buf := make([]byte, packetLen)
	for i := uint32(0); i < packetLen; { // 这里是关键，需要一个循环来处理
		n, err := reader.Read(buf[i:]) // 每次都从未读取的部分开始读取
		if err != nil {
			return []byte(""), err
		}
		i += uint32(n) // 累加已读取的长度
	}

	return buf, nil
}

// 数据解析
func (t *JsonRpcServer) unpack2(reader *bufio.Reader) (string, error) {
	// 对字符串截取，长度数据
	lenByte, _ := reader.Peek(4)
	lengthBuff := bytes.NewBuffer(lenByte) // 建立缓冲区对数据进行读取
	var length int32
	err := binary.Read(lengthBuff, binary.BigEndian, &length)
	if err != nil {
		return "", err
	}
	/*
	   func Read(r io.Reader, order ByteOrder, data interface{}) error
	   从r中读取binary编码的数据并赋给data，data必须是一个指向定长值的指针或者定长值的切片。从r读取的字节使用order指定的字节序解码并写入data的字段里当写入结构体是，名字中有'_'的字段会被跳过，这些字段可用于填充（内存空间）。
	   第二个参数为要转换成的进制
	   **/

	// 读取数据，读取的为4+msgLen，就是包长度和内容一起读取
	if int32(reader.Buffered()) < length+4 {
		return "", err
	}

	//  真正读取
	pack := make([]byte, int(4+length)) // 创建一个切片，用于存储读取的数据，利用切片的长度去限定读取的长度??
	_, err = reader.Read(pack)
	if err != nil {
		return "", err
	}

	return string(pack[4:]), nil
}

func (t *JsonRpcServer) recover(ctx *context2.RunContext, conn net.Conn) {
	id, idOk := ctx.Get(RPC_ID_FIELD)
	if !idOk {
		id = ""
	}
	defer func() {
		if err := recover(); err != nil {
			log.ErrHandle(err)
			conn.Close()
		}
	}()

	err := recover()
	if err == nil {
		return
	}

	switch err.(type) {
	case exception.MsgException:
		var ex = err.(exception.MsgException)
		data := t.buildErrorJson(jsonRpcHelper.NewError(ex.Code, ex.Msg), id)
		t.sendResponse(conn, data)
		return
	case exception.ErrorException:
		var ex = err.(exception.ErrorException)
		log.ErrHandle(err)
		data := t.buildErrorJson(jsonRpcHelper.NewError(ex.Code, ex.Msg), id)
		t.sendResponse(conn, data)
		return
	default:
		log.ErrHandle(err)
		data := t.buildErrorJson(jsonRpcHelper.NewError(jsonRpcHelper.CodeInternalError, ""), id)
		t.sendResponse(conn, data)
		return
	}

}

func (t *JsonRpcServer) sendResponse(conn net.Conn, res []byte) {
	// 先将长度作为header
	returnlenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(returnlenBuf, uint32(len(res)))

	// 拼接长度和内容
	data := append(returnlenBuf, res...)
	//fmt.Println(string(d), len(d), len(data))

	// 发送数据
	_, err := conn.Write(data)
	if err != nil {
		di.Zap().Errorf("conn write error: %v", err)
	}
}

func (t *JsonRpcServer) buildErrorJson(error *jsonRpcHelper.Error, id any) []byte {
	data := map[string]any{
		"jsonrpc": "2.0",
		"error":   error,
		"id":      id,
	}
	json, err := di.Json().Marshal(data)
	if err != nil {
		panic(err)
	}
	return json
}

func (t *JsonRpcServer) buildSuccessJson(data any, id any) []byte {
	res := map[string]any{
		"jsonrpc": "2.0",
		"data":    data,
		"id":      id,
	}
	json, err := di.Json().Marshal(res)
	if err != nil {
		panic(err)
	}
	return json
}

func (t *JsonRpcServer) dispatch(ctx *context2.RunContext, received []byte) (id any, resError *jsonRpcHelper.Error, resData any) {
	var receivedData map[string]any
	id = ""
	if err := di.Json().Unmarshal(received, &receivedData); err != nil {
		di.Zap().Debugf("json parsing error, %v", err)
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeParseError, "")
		return
	}
	if _, ok := receivedData["id"]; !ok {
		di.Zap().Debug("undefined id")
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, "")
		return
	}

	id = receivedData["id"]
	ctx.Set(RPC_ID_FIELD, id)

	if _, ok := receivedData["jsonrpc"]; !ok {
		di.Zap().Debug("undefined jsonrpc")
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, "")
		return
	}

	if receivedData["jsonrpc"] != "2.0" {
		di.Zap().Debug("jsonrpc is not 2.0")
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, "")
		return
	}

	if _, ok := receivedData["method"]; !ok {
		di.Zap().Debug("method")
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, "")
		return
	}

	callStr := receivedData["method"].(string)
	callArr := strings.Split(callStr, ".")

	if len(callArr) != 2 {
		di.Zap().Debug("method parsing error")
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, "")
		return
	}

	var serviceName string
	serviceName = callArr[0]
	if _, ok := t.serviceMap[callArr[0]]; !ok {
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, "")
		return
	}
	var method string
	var methodParams any
	methodParamsMap := make(map[string]any)
	var methodParamsOk bool
	methodParams, methodParamsOk = receivedData["params"]
	if methodParamsOk {
		if methodParams != "" && methodParams != nil {
			var methodParamsMapOk bool
			if methodParamsMap, methodParamsMapOk = methodParams.(map[string]any); !methodParamsMapOk {
				resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, "params必须是对象")
				return
			}
		}
	}

	method = callArr[1]
	params := []any{ctx, methodParamsMap}

	var callResult []reflect.Value
	callResult, resError = t.callMethod(serviceName, callStr, t.serviceMap[callArr[0]], method, params)
	if resError != nil {
		return
	}
	if len(callResult) != 2 {
		di.Zap().Errorf("result len error, result: %v", callResult)
		resError = jsonRpcHelper.NewError(jsonRpcHelper.CodeInternalError, "")
		return
	}

	resError = callResult[1].Interface().(*jsonRpcHelper.Error)
	if resError != nil {
		return
	}

	resData = callResult[0].Interface()

	return id, nil, resData
}

func (t *JsonRpcServer) callMethod(serviceName string, callStr string, obj any, funcName string, params []any) ([]reflect.Value, *jsonRpcHelper.Error) {
	var structValue reflect.Value
	if _, ok := t.structValueMap[serviceName]; !ok {
		structValue = reflect.ValueOf(obj)
	} else {
		structValue = t.structValueMap[serviceName]
	}

	// 检查方法是否存在
	var method reflect.Value
	if _, ok := t.callValueMap[callStr]; ok {
		method = t.callValueMap[callStr]
	} else {
		method = structValue.MethodByName(funcName)
		if !method.IsValid() {
			return nil, jsonRpcHelper.NewError(jsonRpcHelper.CodeInvalidRequest, callStr+" not found")
		}
	}

	inputs := make([]reflect.Value, len(params))
	for i := range params {
		inputs[i] = reflect.ValueOf(params[i])
	}

	result := method.Call(inputs)
	return result, nil
}

func (t *JsonRpcServer) Test(msg string) {
	server := NewJsonRpcServer()
	server.Register("Hello", &service.HelloService{})
	ctx := context2.NewRunContext()
	fmt.Println(server.dispatch(ctx, []byte(msg)))
}
