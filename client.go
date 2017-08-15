package sockets

import (
	"github.com/wzshiming/ffmt"
)

// 客户状态
type Client struct {
	s     *Sockets
	Conn  Conn
	Ident string
	Session
}

// 发送信息
func (c *Client) Send(method string, data *Data) error {
	d, err := NewMessage(method, data).Json()
	if err != nil {
		return err
	}
	err = c.Conn.Write(d)
	if err != nil {
		return err
	}
	return nil
}

// 监听消息
func (c *Client) Listen(f func(string, *Data)) {
	for !c.Conn.IsClose() {
		d, err := c.Conn.Read()
		if err != nil {
			ffmt.Mark(err)
			break
		}

		msg := &Message{}
		err = msg.Unjson(d)
		if err != nil {
			ffmt.Mark(err, string(d))
			break
		}

		f(msg.Method, msg.Data)
	}
	c.Close()

}

// 关闭连接
func (c *Client) Close() error {
	defer c.s.Del(c.Ident)
	if !c.Conn.IsClose() {
		return c.Conn.Close()
	}
	return nil
}
