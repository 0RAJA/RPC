package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

/*
golang写RPC程序，必须符合4个基本条件，不然RPC用不了
	结构体字段首字母要大写，可以别人调用
	函数名必须首字母大写
	函数第一参数是接收参数，第二个参数是返回给客户端的参数，必须是指针类型
	函数还必须有一个返回值error
*/

type Params struct {
	A, B float64
	Opt  string
}

type Response struct {
	Ret float64
}

type Operating struct {
}

func (o *Operating) CallOption(params Params, response *Response) error {
	switch params.Opt {
	case "+":
		response.Ret = params.A + params.B
	case "-":
		response.Ret = params.A - params.B
	case "*":
		response.Ret = params.A * params.B
	case "/":
		if params.B == 0 {
			return errors.New("0作为除数")
		}
		response.Ret = params.A / params.B
	default:
		return errors.New("未知操作")
	}
	return nil
}

func main() {
	option := new(Operating)
	rpc.Register(option)
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go process(conn)
	}
}

func process(conn net.Conn) {
	fmt.Println("new Client")
	jsonrpc.ServeConn(conn)
}
