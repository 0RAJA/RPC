package rpc

import (
	"encoding/binary"
	"io"
	"net"
)

//网络传输数据格式
//两端要约定好数据包的格式
//4字节header uint32

type Session struct {
	conn net.Conn
}

func NewSession(conn net.Conn) *Session {
	return &Session{conn: conn}
}

func (s *Session) Write(data []byte) error {
	buf := make([]byte, 4+len(data))
	//写入头部，记录数据长度
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data)))
	copy(buf[4:], data)
	_, err := s.conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) Read() ([]byte, error) {
	header := make([]byte, 4)
	//读头
	_, err := io.ReadFull(s.conn, header)
	if err != nil {
		return nil, err
	}
	//读数据
	dataLen := binary.BigEndian.Uint32(header)
	data := make([]byte, dataLen)
	if _, err := io.ReadFull(s.conn, data); err != nil {
		return nil, err
	}
	return data, nil
}
