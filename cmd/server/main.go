package main

import (
	"flag"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/pkg/token"
	"github.com/0RAJA/RPC/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

func main() {
	portPtr := flag.Int("port", 8080, "server port")
	flag.Parse()
	log.Println("server port:", *portPtr)
	userStore := service.NewInMemoryUserStoreStore()
	maker, err := token.NewPasetoMaker([]byte(Secret))
	if err != nil {
		log.Fatalln("create maker error:", err)
	}
	//初始化用户存储
	if err := seedUsers(userStore); err != nil {
		log.Fatalln("cannot seed user store:", err)
	}
	authServer := service.NewAuthServer(userStore, maker)
	//初始化拦截器
	interceptor := service.NewAuthInterceptor(maker, accessibleRoles())

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img", ImageMaxSize)
	scoreStore := service.NewInMemoryRateStoreStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, scoreStore)
	grpcServer := grpc.NewServer(
		//安装一个一元拦截器
		grpc.UnaryInterceptor(interceptor.Unary()),
		//安装一个流拦截器
		grpc.StreamInterceptor(interceptor.Stream()),
	)
	reflection.Register(grpcServer) //注册反射服务
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

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
