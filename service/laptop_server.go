package service

import (
	"bytes"
	"errors"
	"github.com/0RAJA/RPC/pb"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

type LaptopServer struct {
	laptopStore LaptopStore //电脑存储
	imageStore  ImageStore  //图像存储
	rateStore   RateStore   //评分存储
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, rateStore RateStore) *LaptopServer {
	return &LaptopServer{laptopStore: laptopStore, imageStore: imageStore, rateStore: rateStore}
}

func (laptop *LaptopServer) LaptopStore() LaptopStore {
	return laptop.laptopStore
}

func (laptop *LaptopServer) ImageStore() ImageStore {
	return laptop.imageStore
}

func (laptop *LaptopServer) RateStore() RateStore {
	return laptop.rateStore
}

// CreateLaptop 创建一个laptop
func (laptop *LaptopServer) CreateLaptop(ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	log.Println("received CreateLaptop request with id:", req.Laptop.Id)
	if len(req.Laptop.Id) > 0 {
		//解析uuid
		_, err := uuid.Parse(req.Laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not valid UUID:%v", err)
		}
	} else {
		id, err := uuid.NewUUID()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID:%v", err)
		}
		req.Laptop.Id = id.String()
	}
	//do something with timeout
	//time.Sleep(time.Second * 6)
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	//保存Laptop
	err := laptop.laptopStore.Save(req.Laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop:%v", err)
	}
	log.Printf("Save laptop with id:%s\n", req.Laptop.Id)
	res := &pb.CreateLaptopResponse{Id: req.Laptop.Id}
	return res, nil
}

// SearchLaptop 查找指定条件的Laptop
func (laptop *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.Filter
	log.Println("received filter:", filter)

	if err := laptop.LaptopStore().Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
		if err := stream.Send(&pb.SearchLaptopResponse{Laptop: laptop}); err != nil {
			return err
		}
		log.Println("send Laptop response with id:", laptop.Id)
		return nil
	}); err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}
	return nil
}

// UploadLaptop 上传图片
func (laptop *LaptopServer) UploadLaptop(stream pb.LaptopService_UploadLaptopServer) error {
	req, err := stream.Recv() //接收第一个信息
	if err != nil {
		return logErr(status.Errorf(codes.Unknown, "can't receive image info:%v", err))
	}
	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("reviving image server,laptopID:%v,imageType:%v", laptopID, imageType)
	//需要确保laptopID存在
	fLaptop, err := laptop.LaptopStore().Find(laptopID)
	if err != nil {
		return logErr(status.Errorf(codes.Internal, "can not find laptop"))
	}
	if fLaptop == nil {
		return logErr(status.Errorf(codes.NotFound, "cannot not found laptop"))
	}
	//可以接收其他的部分了
	imageData := bytes.Buffer{}
	var imageSize uint64

	for {
		//超时判断
		if err := contextErr(stream.Context()); err != nil {
			return err
		}
		//time.Sleep(time.Second * 3)
		log.Println("waiting for receive data")
		req, err := stream.Recv()
		if err != nil {
			log.Println("no more data received")
			break
		}
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err.Error()))
		}
		log.Println("received chunk data with size: ", imageSize)
		chunk := req.GetChunkData()
		size := len(chunk)
		//判断大小
		imageSize += uint64(size)
		if imageSize > laptop.ImageStore().MaxSize() {
			return logErr(status.Errorf(codes.InvalidArgument, "image size is larger than:%v", laptop.imageStore.MaxSize()))
		}
		//写入数据
		if _, err := imageData.Write(chunk); err != nil {
			return logErr(status.Errorf(codes.Internal, "can't write chunk:%v", err))
		}
	}
	imageID, err := laptop.ImageStore().Save(laptopID, imageType, imageData)
	if err != nil {
		return logErr(status.Errorf(codes.Internal, "can't save image:%v", err))
	}
	if err := stream.SendAndClose(&pb.UploadLaptopResponse{
		Id:   imageID,
		Size: imageSize,
	}); err != nil {
		return logErr(status.Errorf(codes.Internal, "can`t send and close:%v", err))
	}
	return nil
}

// RateLaptop 处理评分
func (laptop *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		if err := contextErr(stream.Context()); err != nil {
			return err
		}
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return logErr(status.Errorf(codes.Unknown, "cannot receive stream request:%v", err))
		}
		laptopID := req.GetLaptopId()
		score := req.GetScore()
		log.Printf("received a laptop-score id:%v,score:%v\n", laptopID, score)
		//检查是否存在对应laptopID
		found, err := laptop.LaptopStore().Find(laptopID)
		if err != nil {
			return logErr(status.Errorf(codes.Internal, "cannot find laptop:%v", err))
		}
		if found == nil {
			return logErr(status.Errorf(codes.NotFound, "can not find laptop"))
		}
		rating, err := laptop.RateStore().Add(laptopID, score)
		if err != nil {
			return logErr(status.Errorf(codes.Internal, "store error: %v", err))
		}
		res := &pb.RateLaptopResponse{LaptopId: laptopID, RateCount: rating.Count, AverageScore: rating.Sum / float64(rating.Count)}
		if err := stream.Send(res); err != nil {
			return logErr(status.Errorf(codes.Unknown, "cannot send response:%v", err))
		}
	}
	return nil
}

//简化日志打印
func logErr(err error) error {
	log.Println(err)
	return err
}

func contextErr(ctx context.Context) error {
	switch ctx.Err() {
	case context.DeadlineExceeded:
		return logErr(status.Error(codes.DeadlineExceeded, "deadline exceeded"))
	case context.Canceled:
		return logErr(status.Errorf(codes.Canceled, "canceled"))
	default:
		return nil
	}
}
