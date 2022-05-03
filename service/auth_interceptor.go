package service

import (
	"github.com/0RAJA/RPC/pkg/token"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
)

const Authorization = "authorization"

//权限校验

type AuthInterceptor struct {
	jwtMaker        token.Maker
	accessibleRoles map[string][]string //RPC对应的Roles
}

func NewAuthInterceptor(jwtMaker token.Maker, accessibleRoles map[string][]string) *AuthInterceptor {
	return &AuthInterceptor{jwtMaker: jwtMaker, accessibleRoles: accessibleRoles}
}

// Unary 一元拦截器
func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		log.Println("-->unary Interceptor: ", info.FullMethod)
		if err := interceptor.authorized(ctx, info.FullMethod); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// Stream 流式拦截器
func (interceptor *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Println("-->stream Interceptor: ", info.FullMethod)
		if err := interceptor.authorized(ss.Context(), info.FullMethod); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func (interceptor *AuthInterceptor) authorized(ctx context.Context, method string) error {
	accessibleRoles, ok := interceptor.accessibleRoles[method]
	if !ok {
		//没有设置拦截
		return nil
	}
	//从ctx中获取访问信息
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}
	values := md[Authorization]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}
	//存储在第一个位置
	accessToken := values[0]
	payload, err := interceptor.jwtMaker.VerifyToken(accessToken)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "access token is not valid:%v", err)
	}
	for _, role := range accessibleRoles {
		if payload.Role == role {
			return nil
		}
	}
	return status.Errorf(codes.PermissionDenied, "no permissions to access this rpc")
}
