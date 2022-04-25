package rpc

import (
	"errors"
	"log"
	"net"
	"reflect"
)

/*
服务端接收到的数据需要包括什么？
	调用的函数名、参数列表，还有一个返回值error类型
服务端需要解决的问题是什么？
	Map维护客户端传来调用函数，服务端知道去调谁
服务端的核心功能有哪些？
	维护函数map
	客户端传来的东西进行解析
	函数的返回值打包，传给客户端
*/

type Server struct {
	//地址
	addr string
	//map Map维护客户端传来调用函数，服务端知道去调谁
	funcs map[string]reflect.Value
}

func NewServer(addr string) *Server {
	return &Server{addr: addr, funcs: make(map[string]reflect.Value)}
}

// Register 注册服务
// 第一个参数函数名，第二个传入真正的函数
func (s *Server) Register(rpcName string, f interface{}) {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		panic(errors.New("invalid TypeOf Func"))
	}
	if _, ok := s.funcs[rpcName]; ok {
		panic(errors.New("repeated rpcName"))
	}
	s.funcs[rpcName] = v
}

// Run 服务器等待调用的方法
func (s *Server) Run() {
	listen, err := net.Listen("tcp", s.addr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}
		go s.Process(conn)
	}
}

func (s *Server) Process(conn net.Conn) {
	session := NewSession(conn)
	defer session.conn.Close()
	b, err := session.Read()
	if err != nil {
		log.Println("Error reading session,err:", err)
		return
	}
	rpcData := new(RPCData)
	if err := rpcData.Decode(b); err != nil {
		log.Println("Error decoding session,err:", err)
		return
	}
	f, ok := s.funcs[rpcData.Name]
	if !ok {
		log.Println("Err Call Func Not Find,func name:", rpcData.Name)
		return
	}
	t := f.Type()
	if t.NumIn() != len(rpcData.Args) {
		log.Println("Err NumIn Lens Not Equal")
		return
	}
	inArgs := make([]reflect.Value, 0, len(rpcData.Args))
	for i, arg := range rpcData.Args {
		v := reflect.ValueOf(arg)
		if t.In(i).Kind() != v.Type().Kind() {
			log.Println("Err parameter Type Mismatch,need:", t.In(i).Kind(), "but:", v.Type().Kind())
			return
		}
		inArgs = append(inArgs, v)
	}
	out := f.Call(inArgs)
	outArgs := make([]interface{}, 0, len(out))
	for _, o := range out {
		outArgs = append(outArgs, o.Interface())
	}
	responRPCData := NewRPCData(rpcData.Name, outArgs)
	bytes, err := responRPCData.Encode()
	if err != nil {
		log.Println("Encode error,err:", err)
		return
	}
	if err := session.Write(bytes); err != nil {
		log.Println("Send error,err:", err)
	}
	return
}
