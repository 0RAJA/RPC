package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

const ImageMaxSize = 1 << 20 //1MB

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}

func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log.Println("-->unary Interceptor: ", info.FullMethod)
	return handler(ctx, req)
}

func main() {
	portPtr := flag.Int("port", 8080, "server port")
	flag.Parse()
	log.Println("server port:", *portPtr)
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img", ImageMaxSize)
	scoreStore := service.NewInMemoryRateStoreStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, scoreStore)
	grpcServer := grpc.NewServer(
		//安装一个一元拦截器
		grpc.UnaryInterceptor(unaryInterceptor),
		//安装一个流拦截器
	)
	reflection.Register(grpcServer) //注册反射服务
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	lister, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *portPtr))
	if err != nil {
		log.Fatalln("cannot listen on port,err:", err)
	}
	if err := grpcServer.Serve(lister); err != nil {
		log.Fatalln("cannot start server,err:", err)
	}
}
