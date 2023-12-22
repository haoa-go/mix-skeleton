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
	config Config
	pool   pool.Pool
}

type Config struct {
	Network       string
	Address       string
	InitialCap    int
	MaxCap        int
	ConnTimeOut   int
	ReadWaitTime  int
	WriteWaitTime int
	IdBuilder     func() any
}

func NewClient(c Config) *Client {
	factory := func() (net.Conn, error) {
		return net.DialTimeout(c.Network, c.Address, time.Duration(c.ConnTimeOut)*time.Second)
	}
	if c.InitialCap == 0 {
		c.InitialCap = 5
	}
	if c.MaxCap == 0 {
		c.MaxCap = 30
	}
	//if c.IdBuilder == nil {
	//	panic(errors.New("IdBuilder can not be nil"))
	//}
	p, err := pool.NewChannelPool(c.InitialCap, c.MaxCap, factory)
	if err != nil {
		panic(err)
	}
	return &Client{
		config: c,
		pool:   p,
	}
}

func (t *Client) CallWithTime(id any, method string, params map[string]any, readWaitTime, writeWaitTime int) []byte {
	var conn net.Conn
	var poolGetErr error

	conn, poolGetErr = t.pool.Get()
	if poolGetErr != nil {
		panic(poolGetErr)
	}
	defer conn.Close()

	callData := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      id,
	}
	json, err := di.Json().Marshal(callData)
	if err != nil {
		panic(err)
	}

	sendErr := t.send(conn, json, writeWaitTime)
	if sendErr != nil {
		t.closeConn(conn)
		conn = t.getNewConn()
		defer conn.Close()
		sendErr2 := t.send(conn, json, writeWaitTime)
		if sendErr2 != nil {
			panic(sendErr2)
		}
	}

	data, readErr := t.read(conn, readWaitTime)
	if readErr != nil {
		panic(readErr)
	}
	return data
}

func (t *Client) Call(id any, method string, params map[string]any) []byte {
	return t.CallWithTime(id, method, params, 0, 0)
}

func (t *Client) getNewConn() net.Conn {
	conn, err := net.Dial(t.config.Network, t.config.Address)
	if err != nil {
		panic(conn)
	}
	return conn
}

func (t *Client) closeConn(conn net.Conn) {
	if pc, ok := conn.(*pool.PoolConn); ok {
		pc.MarkUnusable()
		pc.Close()
	} else {
		conn.Close()
	}
}

func (t *Client) send(conn net.Conn, content []byte, waitTime int) error {
	if waitTime == 0 {
		waitTime = t.config.WriteWaitTime
	}
	if waitTime == 0 {
		waitTime = 5
	}
	if err := conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(waitTime))); err != nil {
		return err
	}
	// 先将长度作为header
	returnlenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(returnlenBuf, uint32(len(content)))

	// 拼接长度和内容
	data := append(returnlenBuf, content...)
	//fmt.Println(string(d), len(d), len(data))

	// 发送数据
	_, err := conn.Write(data)
	if err != nil {
		if !strings.Contains(err.Error(), "write: broken pipe") {
			di.Zap().Errorf("conn write error: %v", err)
		} else {
			di.Zap().Debugf("conn write error: %v", err)
		}
		return err
	}
	return nil
}

func (t *Client) read(conn net.Conn, waitTime int) ([]byte, error) {
	if waitTime == 0 {
		waitTime = t.config.ReadWaitTime
	}
	if waitTime == 0 {
		waitTime = 5
	}
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(waitTime))); err != nil {
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
