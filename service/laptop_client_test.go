package service_test

import (
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"github.com/0RAJA/RPC/serializer"
	"github.com/0RAJA/RPC/service"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func checkLaptopSame(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := serializer.ProtobufToJson(laptop1)
	require.NoError(t, err)
	json2, err := serializer.ProtobufToJson(laptop2)
	require.NoError(t, err)
	require.Equal(t, json1, json2)
}

//测试从客户端请求RPC
func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()
	laptopServer, addr := startTestLaptopServer(t)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	LaptopClient := pb.NewLaptopServiceClient(conn)
	sampleLaptop := sample.NewLaptop()
	res, err := LaptopClient.CreateLaptop(context.Background(), &pb.CreateLaptopRequest{Laptop: sampleLaptop})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, res.Id, sampleLaptop.Id)

	//测试Store是否存储
	other, err := laptopServer.Store().Find(sampleLaptop.Id)
	require.NoError(t, err)
	//测试是否相同
	checkLaptopSame(t, sampleLaptop, other)
}

//启动测试服务器
func startTestLaptopServer(t *testing.T) (*service.LaptopServer, string) {
	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	listenner, err := net.Listen("tcp", ":0") //随机分配IP
	require.NoError(t, err)
	go grpcServer.Serve(listenner) //防止阻塞
	return laptopServer, listenner.Addr().String()
}
