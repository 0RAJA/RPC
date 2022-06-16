package main

import (
	"encoding/gob"
	"fmt"
	"log"

	"github.com/0RAJA/RPC/test1/rpc/rpc"
	ss "github.com/0RAJA/RPC/test1/rpc/rpc/cmd/struct"
)

func SayHello(user ss.User) string {
	return fmt.Sprintf("hello %s:%d", user.Name, user.Age)
}

func main() {
	log.Println("server starting...")
	gob.Register(ss.User{})
	srv := rpc.NewServer(":8080")
	srv.Register("SayHello", SayHello)
	srv.Run()
}
