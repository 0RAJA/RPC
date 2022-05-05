package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/pkg/token"
	"github.com/0RAJA/RPC/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"log"
	"net"
)

const (
	ImageMaxSize = 1 << 20 //1MB
	Secret       = "12345678123456781234567812345678"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}

//规则表,仅将需要进行认证的接口放到这里
func accessibleRoles() map[string][]string {
	const LaptopServiceBasePath = "/rpc.proto.LaptopService/"

	return map[string][]string{
		LaptopServiceBasePath + "CreateLaptop": {"admin"},
		LaptopServiceBasePath + "UploadLaptop": {"admin"},
		LaptopServiceBasePath + "RateLaptop":   {"admin", "user"},
	}
}

//加载TLS凭证
func loadTLSCredentials() (credentials.TransportCredentials, error) {
	//服务器端TLS，需要加载服务器的证书和私钥
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, err
	}
	//创建CA证书池并添加证书
	certPool := x509.NewCertPool()
	pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}
	//创建传输凭证,使用服务器证书制作一个tls配置对象
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert, // tls.NoClientCert 服务端TLS
		ClientCAs:    certPool,
	}
	return credentials.NewTLS(config), nil
}

func main() {
	portPtr := flag.Int("port", 8080, "server port")
	enableTLSPtr := flag.Bool("tls", false, "enable tls") //是否开启TLS
	flag.Parse()
	log.Println("server port:", *portPtr, " tls=", *enableTLSPtr)
	//初始化持久层
	userStore := service.NewInMemoryUserStoreStore()
	maker, err := token.NewPasetoMaker([]byte(Secret))
	if err != nil {
		log.Fatalln("create maker error:", err)
	}
	//初始化用户存储
	if err := seedUsers(userStore); err != nil {
		log.Fatalln("cannot seed user store:", err)
	}
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img", ImageMaxSize)
	scoreStore := service.NewInMemoryRateStoreStore()

	laptopServer := service.NewLaptopServer(laptopStore, imageStore, scoreStore)

	authServer := service.NewAuthServer(userStore, maker)

	//初始化拦截器
	interceptor := service.NewAuthInterceptor(maker, accessibleRoles())
	serviceOptions := []grpc.ServerOption{
		//安装一个一元拦截器
		grpc.UnaryInterceptor(interceptor.Unary()),
		//安装一个流拦截器
		grpc.StreamInterceptor(interceptor.Stream()),
	}
	//添加TLS凭证
	if *enableTLSPtr {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			log.Fatalln("can't load TLS credentials:", err)
		}
		serviceOptions = append(serviceOptions, grpc.Creds(tlsCredentials))
	}
	//配置gRPC服务器
	grpcServer := grpc.NewServer(
		serviceOptions...,
	)
	reflection.Register(grpcServer)                          //注册反射服务
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer) //注册laptop服务
	pb.RegisterAuthServiceServer(grpcServer, authServer)     //注册auth服务

	lister, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *portPtr))
	if err != nil {
		log.Fatalln("cannot listen on port,err:", err)
	}
	if err := grpcServer.Serve(lister); err != nil {
		log.Fatalln("cannot start server,err:", err)
	}
}

func seedUsers(userStore service.UserStore) error {
	if err := createUser(userStore, "admin1", "secret", "admin"); err != nil {
		return err
	}
	if err := createUser(userStore, "user1", "secret", "user"); err != nil {
		return err
	}
	return nil
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return fmt.Errorf("createUser err: %v", err)
	}
	if err := userStore.Save(user); err != nil {
		return fmt.Errorf("save user err: %v", err)
	}
	return nil
}
