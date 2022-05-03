package main

import (
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/client"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"google.golang.org/grpc"
	"log"
	"strings"
	"time"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}

const (
	username        = "admin1"
	password        = "secret"
	refreshDuration = 30 * time.Second
)

//规则表,仅将需要进行认证的接口放到这里
func accessibleRoles() map[string]bool {
	const LaptopServiceBasePath = "/rpc.proto.LaptopService/"

	return map[string]bool{
		LaptopServiceBasePath + "CreateLaptop": true,
		LaptopServiceBasePath + "UploadLaptop": true,
		LaptopServiceBasePath + "RateLaptop":   true,
	}
}

func main() {
	serverAddrPtr := flag.String("addr", ":8080", "the server address")
	flag.Parse()
	log.Println("dial server address: ", *serverAddrPtr)

	conn1, err := grpc.Dial(*serverAddrPtr, grpc.WithInsecure()) //用于auth验证的连接
	if err != nil {
		log.Fatalln("cannot connect to server:", err)
	}
	authClient := client.NewAuthClient(conn1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, accessibleRoles(), refreshDuration)
	if err != nil {
		log.Fatalln("can't create auth interceptor:", err)
	}
	conn2, err := grpc.Dial(*serverAddrPtr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(interceptor.Unary()), grpc.WithStreamInterceptor(interceptor.Stream())) //用于auth验证的连接
	if err != nil {
		log.Fatalln("cannot connect to server:", err)
	}
	laptopClient := client.NewLaptopClient(conn2)
	//testCreateLaptop(laptopClient)
	//testSearchLaptop(laptopClient)
	//testUpdateImage(laptopClient)
	testRateLaptop(laptopClient)
}

func testUpdateImage(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.Id, "tmp/test.jpg")
}

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(sample.NewLaptop())
	}
	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_MEGABYTE},
	}
	laptopClient.SearchLaptop(filter)
}

func testRateLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIDs := make([]string, n)
	for i := 0; i < 3; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		laptopClient.CreateLaptop(laptop)
	}
	scores := make([]float64, n)
	for {
		fmt.Println("rate laptop y/n?")
		var ans string
		fmt.Scan(&ans)
		if strings.ToLower(ans) != "y" {
			break
		}
		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}
		if err := laptopClient.RateLaptop(laptopIDs, scores); err != nil {
			log.Fatalln("rateLaptop error: ", err)
		}
	}
}
