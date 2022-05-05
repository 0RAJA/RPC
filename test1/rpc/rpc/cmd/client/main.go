package main

import (
	"encoding/gob"
	"fmt"
	"github.com/0RAJA/RPC/test1/rpc/rpc"
	ss "github.com/0RAJA/RPC/test1/rpc/rpc/cmd/struct"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Fatalln("can't dial: ", err)
	}
	gob.Register(ss.User{})
	cli := rpc.NewClient(conn)
	var SayHello func(user ss.User) string
	errChan := cli.CallRPC("SayHello", &SayHello)
	defer close(errChan)
	res := SayHello(ss.User{Name: "raja", Age: 20})
	fmt.Println(res, <-errChan)
	res = SayHello(ss.User{Name: "ww", Age: 21})
	fmt.Println(res, <-errChan)
}
