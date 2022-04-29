package main

import (
	"flag"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
