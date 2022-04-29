package main

import (
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	portPtr := flag.Int("port", 8080, "server port")
	flag.Parse()
	log.Println("server port:", *portPtr)
	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	lister, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *portPtr))
	if err != nil {
		log.Fatalln("cannot listen on port,err:", err)
	}
	if err := grpcServer.Serve(lister); err != nil {
		log.Fatalln("cannot start server,err:", err)
	}
}
