package rpc

import (
	"errors"
	"net"
	"reflect"
)

type Client struct {
	conn net.Conn
}

func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn}
}

// CallRPC 实现通用的RPC客户端
// 传入访问的函数名
// fPtr指向的是函数原型
// var select fun xx(User)
// cli.callRPC("selectUser",&select)
func (c *Client) CallRPC(rpcName string, fPtr interface{}) chan error {
	fn := reflect.ValueOf(fPtr).Elem() //获取函数原型
	if fn.Type().Kind() != reflect.Func {
		panic(errors.New("err fPtr Not Func"))
	}
	errChan := make(chan error, 1)
	f := func(args []reflect.Value) []reflect.Value {
		inArgs := make([]interface{}, 0, len(args))
		for _, arg := range args {
			inArgs = append(inArgs, arg.Interface())
		}
		session := NewSession(c.conn)
		rpcData := NewRPCData(rpcName, inArgs)
		b, err := rpcData.Encode()
		if err != nil {
			errChan <- errors.New("Err Encode err:" + err.Error())
			return nil
		}
		if err = session.Write(b); err != nil {
			errChan <- err
			return nil
		}
		respBytes, err := session.Read()
		if err != nil {
			errChan <- err
			return nil
		}
		respRPC := new(RPCData)
		if err := respRPC.Decode(respBytes); err != nil {
			errChan <- err
			return nil
		}
		outArgs := make([]reflect.Value, 0, len(respRPC.Args))
		for i, arg := range respRPC.Args {
			//nil转换
			if arg == nil {
				// reflect.Zero()会返回类型的零值的value
				// .out()会返回函数输出的参数类型
				outArgs = append(outArgs, reflect.Zero(fn.Type().Out(i)))
				continue
			}
			outArgs = append(outArgs, reflect.ValueOf(arg))
		}
		errChan <- nil
		return outArgs
	}
	v := reflect.MakeFunc(fn.Type(), f)
	fn.Set(v)
	return errChan
}
