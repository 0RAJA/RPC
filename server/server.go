package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
)
import pb "github.com/0RAJA/RPC/proto"

var (
	port string
)

func init() {
	flag.StringVar(&port, "port", "80", "port")
	flag.Parse()
}

type GreeterServer struct{}

func (s *GreeterServer) SayHello(ctx context.Context, r *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "hello world"}, nil
}

func main() {
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &GreeterServer{})
}
