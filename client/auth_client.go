package client

import (
	"github.com/0RAJA/RPC/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"time"
)

// AuthClient 调用身份验证服务
type AuthClient struct {
	service  pb.AuthServiceClient
	username string
	password string
}

func NewAuthClient(conn *grpc.ClientConn, username, password string) *AuthClient {
	service := pb.NewAuthServiceClient(conn)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

// Login 登陆获取token
func (auth *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := auth.service.Login(ctx, &pb.LoginRequest{Username: auth.username, Password: auth.password})
	if err != nil {
		return "", err
	}
	return res.GetAccessToken(), nil
}
