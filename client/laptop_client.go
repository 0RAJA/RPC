package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/0RAJA/RPC/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LaptopClient struct {
	service pb.LaptopServiceClient
}

func NewLaptopClient(conn *grpc.ClientConn) *LaptopClient {
	return &LaptopClient{service: pb.NewLaptopServiceClient(conn)}
}

func (client *LaptopClient) CreateLaptop(laptop *pb.Laptop) {
	req := &pb.CreateLaptopRequest{Laptop: laptop}

	//设置超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.service.CreateLaptop(ctx, req)
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

func (client *LaptopClient) SearchLaptop(filter *pb.Filter) {
	log.Println("search filter:", filter)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := client.service.SearchLaptop(ctx, req) //总时间超过即超时
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

// UploadImage 上传文件
func (client *LaptopClient) UploadImage(laptopID, imagePath string) error {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalln("cannot open image:", err)
	}
	defer file.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.UploadLaptop(ctx)
	if err != nil {
		log.Fatalln("cannot upload laptop:", err)
	}
	req := &pb.UploadLaptopRequest{
		Data: &pb.UploadLaptopRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId: laptopID, ImageType: filepath.Ext(imagePath),
			},
		},
	}
	if err := stream.Send(req); err != nil {
		log.Fatalln("cannot upload image info:", err)
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("cannot read image info:", err)
		}
		req := &pb.UploadLaptopRequest_ChunkData{ChunkData: buffer[:n]}
		if err := stream.Send(&pb.UploadLaptopRequest{
			Data: req,
		}); err != nil {
			log.Fatalln("cannot send upload:", err)
		}
	}
	rev, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalln("cannot receive response:", err)
	}
	log.Println("success_with_id:", rev.Id, " size:", rev.Size)
	return nil
}

// RateLaptop 创建评分
func (client *LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	stream, err := client.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("can't rate laptop: %v", err)
	}
	//接收并返回结果
	receive := func(ctx context.Context) (<-chan *pb.RateLaptopResponse, <-chan error) {
		errChan := make(chan error)
		resChan := make(chan *pb.RateLaptopResponse)
		go func() {
			defer close(errChan)
			defer close(resChan)
			for {
				res, err := stream.Recv()
				if err == io.EOF {
					break
				}
				var out1, out2 = resChan, errChan //本地版本，隐藏外界变量
				for i := 0; i < 2; i++ {
					select {
					case <-ctx.Done():
						return
					case out1 <- res:
						out1 = nil
					case out2 <- err:
						out2 = nil
					}
				}
				if err != nil {
					return
				}
			}
		}()
		return resChan, errChan
	}
	resStream, errStream := receive(ctx)
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{LaptopId: laptopID, Score: scores[i]}
		if err := stream.Send(req); err != nil {
			return fmt.Errorf("send error: %v", err)
		}
		log.Println("send request laptopID:", req.LaptopId, "score:", req.Score)
		if err := <-errStream; err != nil {
			return fmt.Errorf("err received error: %v", err)
		}
		log.Println("received response:", <-resStream)
	}
	if err := stream.CloseSend(); err != nil {
		return fmt.Errorf("can't close stream: %v", err)
	}
	err = <-errStream
	return err
}
