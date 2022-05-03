package client

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

const Authorization = "authorization"

type AuthInterceptor struct {
	authClient  *AuthClient
	authMethods map[string]bool //role对应的访问方法
	accessToken string
}

// NewAuthInterceptor 返回一个新的拦截器，refreshDuration 明确获取token的时间间隔
func NewAuthInterceptor(authClient *AuthClient, authMethods map[string]bool, refreshDuration time.Duration) (*AuthInterceptor, error) {
	interceptor := &AuthInterceptor{
		authClient:  authClient,
		authMethods: authMethods,
	}
	err := interceptor.scheduleRefreshToken(refreshDuration)
	if err != nil {
		return nil, err
	}
	return interceptor, nil
}

// Unary 拦截器
func (interceptor *AuthInterceptor) Unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Println("--> unary:", method)
		//检验是否需要认证
		if interceptor.authMethods[method] {
			return invoker(interceptor.attachToken(ctx), method, req, reply, cc, opts...)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
func (interceptor *AuthInterceptor) Stream() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log.Println("--> stream:", method)
		if interceptor.authMethods[method] {
			return streamer(interceptor.attachToken(ctx), desc, cc, method, opts...)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

//添加token到context
func (interceptor *AuthInterceptor) attachToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, Authorization, interceptor.accessToken)
}

//自动刷新token
func (interceptor *AuthInterceptor) scheduleRefreshToken(refreshDuration time.Duration) error {
	if err := interceptor.refreshToken(); err != nil {
		return err
	}
	ticker := time.NewTicker(refreshDuration)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := interceptor.refreshToken(); err != nil {
					log.Println("failed to refresh token:", err)
					ticker.Reset(2 * refreshDuration)
				} else {
					ticker.Reset(refreshDuration)
				}
			}
		}
	}()
	return nil
}

//刷新token
func (interceptor *AuthInterceptor) refreshToken() error {
	accessToken, err := interceptor.authClient.Login()
	if err != nil {
		return err
	}
	interceptor.accessToken = accessToken
	log.Println("token refreshed: ", accessToken)
	return nil
}
