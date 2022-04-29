package main

import (
	"context"
	"flag"
	pb "github.com/0RAJA/RPC/test1/proto"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
)

var port string

func init() {
	flag.StringVar(&port, "p", "8080", "启动端口号")
	flag.Parse()
}

type GreeterServer struct{}

// SayRoute 双向流式RPC
func (s *GreeterServer) SayRoute(stream pb.Greeter_SayRouteServer) error {
	n := 0
	for {
		_ = stream.Send(&pb.HelloReply{Message: "say.Route"})
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		n++
		log.Println("n:", n, "resp:", resp)
	}
}

// SayRecord 客户端流式RPC
func (s *GreeterServer) SayRecord(stream pb.Greeter_SayRecordServer) error {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			message := &pb.HelloReply{Message: "OK"}
			return stream.SendAndClose(message) //当流关闭后，将最终响应结果发送给客户端，同时关闭 Recv()
		}
		if err != nil {
			return err
		}
		log.Println(resp.Name)
	}
}

// SayHello 单向RPC
func (s *GreeterServer) SayHello(ctx context.Context, r *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "hello.world"}, nil
}

// SayList 服务端流式RPC
func (s *GreeterServer) SayList(r *pb.HelloRequest, stream pb.Greeter_SayListServer) error {
	for i := 0; i <= 6; i++ {
		_ = stream.Send(&pb.HelloReply{Message: "hello.list"}) //序列化+压缩+header(5bytes) 如果超过MaxInt32报错
	}
	return nil
}

func main() {
	server := grpc.NewServer()                         //创建gRPC Server对象
	pb.RegisterGreeterServer(server, &GreeterServer{}) //注册服务端接口
	lis, err := net.Listen("tcp", ":"+port)            //监听TCP
	if err != nil {
		log.Fatal(err)
	}
	//gRPC Server 开始 lis.Accept
	if err := server.Serve(lis); err != nil {
		panic(err)
	}
}
