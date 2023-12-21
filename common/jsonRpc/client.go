package jsonRpc

import (
	"app/di"
	"bufio"
	"encoding/binary"
	"github.com/fatih/pool"
	"net"
	"strings"
	"time"
)

type Client struct {
	network, address   string
	initialCap, maxCap int
	idBuilder          func() any
	pool               pool.Pool
}

func NewClient(network, address string, initialCap, maxCap int, idBuilder func() any) *Client {
	factory := func() (net.Conn, error) {
		return net.Dial(network, address)
	}
	p, err := pool.NewChannelPool(initialCap, maxCap, factory)
	if err != nil {
		panic(err)
	}
	return &Client{
		network:    network,
		address:    address,
		initialCap: initialCap,
		maxCap:     maxCap,
		idBuilder:  idBuilder,
		pool:       p,
	}
}

func (t *Client) Call(method string, params map[string]any, readWaitTime int) (data []byte, err error) {
	var conn net.Conn
	var connErr error
	conn, connErr = t.pool.Get()
	if connErr != nil {
		return nil, connErr
	}
	defer conn.Close()
	callData := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      t.idBuilder(),
	}
	json, err := di.Json().Marshal(callData)
	if err != nil {
		return nil, err
	}

	sendErr := t.sendResponse(conn, json)
	if sendErr != nil {
		conn, connErr = t.pool.Get()
		if connErr != nil {
			return nil, connErr
		}
		defer conn.Close()
		sendErr2 := t.sendResponse(conn, json)
		if sendErr2 != nil {
			return nil, sendErr2
		}
	}

	data, readErr := t.read(conn, readWaitTime)
	if readErr != nil {
		return nil, readErr
	}

	return data, nil
}

func (t *Client) sendResponse(conn net.Conn, res []byte) error {
	waitTime := 5
	if err := conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(waitTime))); err != nil {
		return err
	}
	// 先将长度作为header
	returnlenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(returnlenBuf, uint32(len(res)))

	// 拼接长度和内容
	data := append(returnlenBuf, res...)
	//fmt.Println(string(d), len(d), len(data))

	// 发送数据
	_, err := conn.Write(data)
	if err != nil {
		if !strings.Contains(err.Error(), "write: broken pipe") {
			di.Zap().Errorf("conn write error: %v", err)
		} else {
			di.Zap().Debugf("conn write error: %v", err)
		}
		if pc, ok := conn.(*pool.PoolConn); ok {
			pc.MarkUnusable()
			pc.Close()
		}
		return err
	}
	return nil
}

func (t *Client) read(conn net.Conn, readWaitTime int) ([]byte, error) {
	if readWaitTime == 0 {
		readWaitTime = 5
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(readWaitTime))); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)

	// 读取数据长度
	lenBuf := make([]byte, 4)
	_, err := reader.Read(lenBuf)
	if err != nil {
		return nil, err
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
