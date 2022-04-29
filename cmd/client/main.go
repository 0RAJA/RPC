package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	serverAddrPtr := flag.String("addr", ":8080", "the server address")
	flag.Parse()
	log.Println("dial server address: ", *serverAddrPtr)

	conn, err := grpc.Dial(*serverAddrPtr, grpc.WithInsecure())
	if err != nil {
		log.Fatalln("cannot connect to server:", err)
	}
	laptopClient := pb.NewLaptopServiceClient(conn)
	//CreateLaptop(laptopClient)
	for i := 0; i < 10; i++ {
		CreateLaptop(laptopClient)
	}
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_MEGABYTE},
	}
	SearchLaptop(laptopClient, filter)
}

func CreateLaptop(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()

	req := &pb.CreateLaptopRequest{Laptop: laptop}

	//设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Println("the laptop already exists")
		} else {
			log.Fatalln("cannot createLaptop:", err)
		}
	} else {
		log.Println("created successfully:", res.Id)
	}
}

func SearchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Println("search filter:", filter)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req) //总时间超过即超时
	if err != nil {
		log.Fatalln("search error:", err)
	}
	for {
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("recv over")
			return
		}
		if err != nil {
			log.Fatalln("cannot recv over:", err)
		}
		fmt.Println(res.Laptop.Id, res.Laptop.Cpu)
	}
}
