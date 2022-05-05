package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/client"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
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

//加载TLS凭证
func loadTLSCredentials() (credentials.TransportCredentials, error) {
	//客户端TLS，需要加载客户端的证书和私钥
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}
	//加载签署服务器的CA的证书，客户端需要验证服务器的真实性
	//创建CA证书池并添加证书
	certPool := x509.NewCertPool()
	pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}
	//创建凭据并返回
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool, //受信任的CA证书
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	return credentials.NewTLS(config), nil
}

func main() {
	serverAddrPtr := flag.String("addr", "127.0.0.1:8080", "the server address")
	flag.Parse()
	log.Println("dial server address: ", *serverAddrPtr)
	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatalln("can't load TLS credentials'")
	}
	//用于auth验证的连接
	conn1, err := grpc.Dial(*serverAddrPtr, grpc.WithTransportCredentials(tlsCredentials))
	if err != nil {
		log.Fatalln("cannot connect to server:", err)
	}
	authClient := client.NewAuthClient(conn1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, accessibleRoles(), refreshDuration)
	if err != nil {
		log.Fatalln("can't create auth interceptor:", err)
	}
	//用于auth验证的连接
	conn2, err := grpc.Dial(*serverAddrPtr, grpc.WithTransportCredentials(tlsCredentials), grpc.WithUnaryInterceptor(interceptor.Unary()), grpc.WithStreamInterceptor(interceptor.Stream()))
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
