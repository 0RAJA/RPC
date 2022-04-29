package service

import (
	"errors"
	"github.com/0RAJA/RPC/pb"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

type LaptopServer struct {
	store LaptopStore //
}

func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{store: store}
}

func (laptop *LaptopServer) Store() LaptopStore {
	return laptop.store
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
	switch ctx.Err() {
	case context.DeadlineExceeded:
		log.Println("DeadlineExceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case context.Canceled:
		log.Println("Canceled")
		return nil, status.Errorf(codes.Canceled, "canceled")
	default:
	}
	//保存Laptop
	err := laptop.store.Save(req.Laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrLaptopAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop:%v", err)
	}
	log.Printf("Save laptop with id:%s\n", req.Laptop.Id)
	res := &pb.CreateLaptopResponse{Id: req.Laptop.Id}
	return res, nil
}

func (laptop *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.Filter
	log.Println("received filter:", filter)

	if err := laptop.Store().Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
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
