package service

import (
	"context"
	"github.com/0RAJA/RPC/pb"
	"github.com/0RAJA/RPC/pkg/token"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// AuthServer 权限管理服务
type AuthServer struct {
	userStore UserStore
	jwtMaker  token.Maker
}

func (auth *AuthServer) Login(ctx context.Context, request *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := auth.userStore.Find(request.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't find user:%v", err)
	}
	if user == nil || !user.IsCorrectPassword(request.Password) {
		return nil, status.Errorf(codes.NotFound, "invalid username/password")
	}
	token, _, err := auth.jwtMaker.CreateToken(user.Username, user.Role, 10*time.Minute)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create token:%v", err)
	}
	return &pb.LoginResponse{AccessToken: token}, nil
}

func NewAuthServer(userStore UserStore, jwtMaker token.Maker) *AuthServer {
	return &AuthServer{
		userStore: userStore,
		jwtMaker:  jwtMaker,
	}
}
