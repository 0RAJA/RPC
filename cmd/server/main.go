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

const ImageMaxSize = 1 << 20 //1MB

func main() {
	portPtr := flag.Int("port", 8080, "server port")
	flag.Parse()
	log.Println("server port:", *portPtr)
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img", ImageMaxSize)
	laptopServer := service.NewLaptopServer(laptopStore, imageStore)
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
