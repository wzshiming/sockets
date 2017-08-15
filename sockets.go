package sockets

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/wzshiming/ffmt"
)

type Handler interface {
	ServeConn(*Data, *Client)
}

type HandlerFunc func(*Data, *Client)

func (f HandlerFunc) ServeConn(data *Data, cli *Client) {
	f(data, cli)
}

// 通信维护池
type Sockets struct {
	mut      sync.RWMutex
	mux      map[string]Handler // 路由
	idents   map[string]*Client
	recovery func(*Client, interface{}) error // 处理错误

}

func NewSockets() *Sockets {
	return &Sockets{
		mux:    map[string]Handler{},
		idents: map[string]*Client{},
		recovery: func(cli *Client, i interface{}) error {
			ffmt.MarkStack(1, ffmt.Sputs(time.Now(), cli, i))
			return nil
		},
	}
}

// 人数
func (s *Sockets) Len() int {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return len(s.idents)
}

// 广播
func (s *Sockets) Broadcast(f func(cli *Client)) {
	s.mut.RLock()
	defer s.mut.RUnlock()
	for _, v := range s.idents {
		f(v)
	}
}

// 处理器
func (s *Sockets) Handler(method string, hand Handler) error {
	s.mux[method] = hand
	return nil
}

// 处理器
func (s *Sockets) HandlerFunc(method string, hand HandlerFunc) error {
	return s.Handler(method, hand)
}

// 开启连接监听线程
func (s *Sockets) listen(cli *Client) error {
	cli.Listen(func(method string, d *Data) {
		defer func() {
			if x := recover(); x != nil {
				s.recovery(cli, x)
			}
		}()

		hand := s.mux[method]
		if hand != nil {
			hand.ServeConn(d, cli)
		} else {
			s.recovery(cli, "调用未定义方法")
		}
	})

	return nil
}

// 获取客户端连接
func (s *Sockets) Get(ident string) *Client {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.idents[ident]
}

// 设置客户端连接
func (s *Sockets) Set(ident string, cli *Client) {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.idents[ident] = cli
}

// 关闭连接
func (s *Sockets) Del(ident string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	delete(s.idents, ident)
	return
}

// 接受连接
func (s *Sockets) Accept(conn Conn, sess Session) error {

	ident := makeident(conn)

	cli := &Client{
		s:       s,
		Ident:   ident,
		Session: sess,
		Conn:    conn,
	}

	s.Set(ident, cli)

	return s.listen(cli)
}

// 根据连接生成标识符
func makeident(i interface{}) string {
	val := reflect.ValueOf(i)
	val = reflect.Indirect(val)
	return fmt.Sprint(val.UnsafeAddr())
}
