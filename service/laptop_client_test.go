package service_test

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/sample"
	"github.com/0RAJA/RPC/serializer"
	"github.com/0RAJA/RPC/service"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
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
	laptopStore := service.NewInMemoryLaptopStore()
	addr := startTestLaptopServer(t, laptopStore, nil)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	LaptopClient := pb.NewLaptopServiceClient(conn)
	sampleLaptop := sample.NewLaptop()
	res, err := LaptopClient.CreateLaptop(context.Background(), &pb.CreateLaptopRequest{Laptop: sampleLaptop})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, res.Id, sampleLaptop.Id)

	//测试Store是否存储
	other, err := laptopStore.Find(sampleLaptop.Id)
	require.NoError(t, err)
	//测试是否相同
	checkLaptopSame(t, sampleLaptop, other)
}

//启动测试服务器返回监听地址
func startTestLaptopServer(t *testing.T, laptopStore service.LaptopStore, imageStore service.ImageStore) string {
	laptopServer := service.NewLaptopServer(laptopStore, imageStore)
	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	listener, err := net.Listen("tcp", ":0") //随机分配IP
	require.NoError(t, err)
	go grpcServer.Serve(listener) //防止阻塞
	return listener.Addr().String()
}

//测试搜索
func TestLaptopServer_SearchLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	addr := startTestLaptopServer(t, service.NewInMemoryLaptopStore(), nil)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	laptopClient := pb.NewLaptopServiceClient(conn)

	exceededIDs := make(map[string]bool)
	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 2100
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 4096, Unit: pb.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			exceededIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 2000
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &pb.Memory{Value: 64, Unit: pb.Memory_GIGABYTE}
			exceededIDs[laptop.Id] = true
		}
		req, err := laptopClient.CreateLaptop(context.Background(), &pb.CreateLaptopRequest{Laptop: laptop})
		require.NoError(t, err)
		require.NotEmpty(t, req)
	}

	log.Println("search filter:", filter)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req) //总时间超过即超时
	require.NoError(t, err)
	found := 0
	for {
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			log.Println("recv over")
			break
		}
		require.NoError(t, err)
		found++
		require.True(t, exceededIDs[res.Laptop.Id])
	}
	require.Equal(t, found, len(exceededIDs))
}

//测试上传图片
func TestClientUploadImage(t *testing.T) {
	t.Parallel()

	testImageFolder := "/home/raja/workspace/go/src/RPC/img"
	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore(testImageFolder, 1<<20)
	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddr := startTestLaptopServer(t, laptopStore, imageStore)
	conn, err := grpc.Dial(serverAddr, grpc.WithInsecure())
	require.NoError(t, err)
	laptopClient := pb.NewLaptopServiceClient(conn)
	imagePath := "/home/raja/workspace/go/src/RPC/tmp/test.jpg"

	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer file.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadLaptop(ctx)
	require.NoError(t, err)
	imageType := filepath.Ext(imagePath)
	req := &pb.UploadLaptopRequest{
		Data: &pb.UploadLaptopRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId: laptop.GetId(), ImageType: imageType,
			},
		},
	}
	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0 //记录上传文件的总大小
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		size += n
		require.NoError(t, err)
		req := &pb.UploadLaptopRequest_ChunkData{ChunkData: buffer[:n]}
		err = stream.Send(&pb.UploadLaptopRequest{
			Data: req,
		})
		require.NoError(t, err)
	}
	rev, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, rev.Id)
	require.EqualValues(t, size, rev.GetSize())

	savedImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, rev.GetId(), imageType)
	require.FileExists(t, savedImagePath)         //判断是否存在
	require.NoError(t, os.Remove(savedImagePath)) //最终删除
}
