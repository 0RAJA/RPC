package main

import (
	"fmt"
	"log"
	"net/rpc/jsonrpc"
)

type Params struct {
	A, B float64
	Opt  string
}

type Response struct {
	Ret float64
}

func main() {
	conn, err := jsonrpc.Dial("tcp", ":8080")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	response := new(Response)
	params := &Params{
		A:   1,
		B:   3,
		Opt: "+",
	}
	if err := conn.Call("Operating.CallOption", params, response); err != nil {
		log.Println(err)
		return
	}
	fmt.Println(response.Ret)
}
