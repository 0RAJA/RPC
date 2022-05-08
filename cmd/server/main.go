package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/pkg/token"
	"github.com/0RAJA/RPC/service"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

const (
	ImageMaxSize = 1 << 20 //1MB
	Secret       = "12345678123456781234567812345678"
	ServerCert   = "cert/server-cert.pem"
	ServerKey    = "cert/server-key.pem"
	CACert       = "cert/ca-cert.pem"
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
	serverCert, err := tls.LoadX509KeyPair(ServerCert, ServerKey)
	if err != nil {
		return nil, err
	}
	//创建CA证书池并添加证书
	certPool := x509.NewCertPool()
	pemServerCA, err := ioutil.ReadFile(CACert)
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

type serverConfig struct {
	laptopServer *service.LaptopServer
	authServer   *service.AuthServer
	enableTLS    bool
	listener     net.Listener
	maker        token.Maker
	grpcEndpoint string
}

func runGRPCServer(config *serverConfig) error {
	//初始化拦截器
	interceptor := service.NewAuthInterceptor(config.maker, accessibleRoles())
	serviceOptions := []grpc.ServerOption{
		//安装一个一元拦截器
		grpc.UnaryInterceptor(interceptor.Unary()),
		//安装一个流拦截器
		grpc.StreamInterceptor(interceptor.Stream()),
	}
	//添加TLS凭证
	if config.enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("can't load TLS credentials,error: %w", err)
		}
		serviceOptions = append(serviceOptions, grpc.Creds(tlsCredentials))
	}
	//配置gRPC服务器
	grpcServer := grpc.NewServer(
		serviceOptions...,
	)
	reflection.Register(grpcServer)                                 //注册反射服务
	pb.RegisterLaptopServiceServer(grpcServer, config.laptopServer) //注册laptop服务
	pb.RegisterAuthServiceServer(grpcServer, config.authServer)     //注册auth服务
	log.Println("server port:", config.listener.Addr().String(), " tls=", config.enableTLS)
	//启动gRPC服务
	return grpcServer.Serve(config.listener)
}

func runRESTServer(config *serverConfig) error {
	mux := runtime.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//设置拨号选项
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
	// gRPC到REST的进程间转换 pb.RegisterAuthServiceHandlerServer
	// gRPC网关 RegisterAuthServiceHandlerFromEndpoint
	if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, config.grpcEndpoint, dialOpts); err != nil {
		return err
	}
	if err := pb.RegisterLaptopServiceHandlerFromEndpoint(ctx, mux, config.grpcEndpoint, dialOpts); err != nil {
		return err
	}
	log.Println("server port:", config.listener.Addr().String(), " tls=", config.enableTLS, "grpcEndpoint=", config.grpcEndpoint)
	if config.enableTLS {
		return http.ServeTLS(config.listener, mux, ServerCert, ServerKey)
	}
	return http.Serve(config.listener, mux)
}

func main() {
	portPtr := flag.Int("port", 8080, "server port")
	enableTLSPtr := flag.Bool("tls", false, "enable tls") //是否开启TLS
	serverType := flag.String("type", "grpc", "type of server(grpc/rest)")
	endPointPtr := flag.String("endpoint", "", "grpc endpoint")
	flag.Parse()
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
	addr := fmt.Sprintf("0.0.0.0:%d", *portPtr)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("listener error: ", err)
	}
	config := &serverConfig{
		laptopServer: laptopServer,
		authServer:   authServer,
		enableTLS:    *enableTLSPtr,
		listener:     listener,
		maker:        maker,
		grpcEndpoint: *endPointPtr,
	}
	if *serverType == "grpc" {
		err = runGRPCServer(config)
	} else {
		err = runRESTServer(config)
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
