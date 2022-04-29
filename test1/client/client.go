package main

import (
	pb "github.com/0RAJA/RPC/test1/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	conn, err := grpc.Dial(":"+"8080", grpc.WithInsecure()) //建立连接
	handleErr(err)
	defer conn.Close()
	client := pb.NewGreeterClient(conn) //创建Greeter客户端
	err = SayHello(client)              //发送RPC请求，等待同步响应。
	handleErr(err)
	err = SayList(client, &pb.HelloRequest{Name: "raja"})
	handleErr(err)
	err = SayRecord(client, &pb.HelloRequest{Name: "raja"})
	handleErr(err)
	err = SayRoute(client, &pb.HelloRequest{Name: "raja"})
	handleErr(err)
}

func SayHello(client pb.GreeterClient) error {
	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "raja"})
	if err != nil {
		return err
	}
	log.Println("client.SayHello resp:", resp.Message)
	return nil
}

func SayList(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, err := client.SayList(context.Background(), r)
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		log.Println(resp.Message)
	}
	return nil
}

func SayRecord(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, _ := client.SayRecord(context.Background())
	for n := 0; n < 6; n++ {
		_ = stream.Send(r)
	}
	resp, err := stream.CloseAndRecv() //和服务端配套使用
	if err != nil {
		return err
	}
	log.Println("resp:", resp)
	return nil
}

func SayRoute(client pb.GreeterClient, r *pb.HelloRequest) error {
	stream, _ := client.SayRoute(context.Background())
	for n := 0; n <= 5; n++ {
		_ = stream.Send(r)
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Println("resp:", resp)
	}
	_ = stream.CloseSend()
	return nil
}
