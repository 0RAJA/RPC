package rpc

import (
	"bytes"
	"encoding/gob"
)

//数据的编解码

// RPCData 定义RPC交互的数据结构
type RPCData struct {
	// 访问的函数
	Name string
	// 访问时的参数
	Args []interface{}
}

func NewRPCData(name string, args []interface{}) *RPCData {
	return &RPCData{
		Name: name,
		Args: args,
	}
}

func (data *RPCData) Encode() ([]byte, error) {
	var buf bytes.Buffer
	bufEnc := gob.NewEncoder(&buf)
	if err := bufEnc.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (data *RPCData) Decode(b []byte) error {
	buf := bytes.NewBuffer(b)
	bufDec := gob.NewDecoder(buf)
	if err := bufDec.Decode(data); err != nil {
		return err
	}
	return nil
}
