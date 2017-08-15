package sockets

import (
	"encoding/json"
	"time"
)

type Data []byte

func NewData(i interface{}) *Data {
	d := &Data{}
	d.Marshal(i)
	return d
}

func (m *Data) Marshal(i interface{}) (err error) {
	(*m), err = json.Marshal(i)
	return
}

func (m *Data) Unmarshal(i interface{}) error {
	return json.Unmarshal(*m, i)
}

func (m *Data) MarshalJSON() ([]byte, error) {
	return *m, nil
}

func (m *Data) UnmarshalJSON(data []byte) error {
	*m = data
	return nil
}

type Message struct {
	Method    string // 方法名
	Data      *Data  // 数据
	Unique    uint64 // 防止重复
	Timestamp int64  // 发送时时间戳
}

var unique uint64 = 0

func NewMessageBytes(method string, data []byte) *Message {
	//unique++
	d := Data(data)
	return &Message{
		Method: method,
		Data:   &d,
		//Unique:    unique,
		Timestamp: time.Now().UnixNano(),
	}
}

func NewMessage(method string, data *Data) *Message {
	//unique++
	return &Message{
		Method: method,
		Data:   data,
		//Unique:    unique,
		Timestamp: time.Now().UnixNano(),
	}
}

func (m *Message) Json() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unjson(data []byte) error {
	return json.Unmarshal(data, m)
}
