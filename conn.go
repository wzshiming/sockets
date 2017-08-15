package sockets

import (
	"io"

	"sync"

	"github.com/wzshiming/ffmt"
	"golang.org/x/net/websocket"
)

// 连接
type Conn interface {
	Write([]byte) error    // 写数据
	Read() ([]byte, error) // 读数据
	Close() error          // 关闭连接
	IsClose() bool         // 是关闭的
}

// conn websocket
type ConnWs struct {
	Conn *websocket.Conn
	rmux sync.Mutex
	wmux sync.Mutex
}

func (c *ConnWs) IsClose() bool {
	return c.Conn == nil
}

func (c *ConnWs) Close() error {
	if c.IsClose() {
		return nil
	}
	defer func() {
		c.Conn = nil
	}()
	return c.Conn.Close()
}

func (c *ConnWs) Write(p []byte) error {
	if c.IsClose() {
		return io.EOF
	}

	c.wmux.Lock()
	defer c.wmux.Unlock()

	w, err := c.Conn.NewFrameWriter(websocket.TextFrame)
	if err != nil {
		return err
	}

	_, err = w.Write(p)
	if err != nil {
		return err
	}

	return w.Close()
}

func (c *ConnWs) Read() ([]byte, error) {
	if c.IsClose() {
		return nil, io.EOF
	}

	c.rmux.Lock()
	defer c.rmux.Unlock()

	r, err := c.Conn.NewFrameReader()
	if err != nil {
		return nil, err
	}

	switch r.PayloadType() {
	case websocket.ContinuationFrame:
		ffmt.Mark("ContinuationFrame")
	case websocket.TextFrame:
		data := make([]byte, r.Len())
		i, err := r.Read(data)
		if err != nil {
			return nil, err
		}
		data = data[:i]
		return data, nil
	case websocket.BinaryFrame:
		ffmt.Mark("BinaryFrame")
	case websocket.CloseFrame:
		ffmt.Mark("CloseFrame")
	case websocket.PingFrame:
		ffmt.Mark("PingFrame")
	case websocket.PongFrame:
		ffmt.Mark("PongFrame")
	default:
		ffmt.Mark("UnknownFrame", r.PayloadType())
	}

	data := make([]byte, r.Len())
	i, err := r.Read(data)
	if err != nil {
		return nil, err
	}
	ffmt.Mark(string(data[:i]))

	c.Close() // 未定义的直接关闭 连接

	return nil, nil
}
