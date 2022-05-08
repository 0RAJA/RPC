# RPC学习项目

<a href="https://www.youtube.com/watch?v=2Sm_O75I7H0">原视频</a>

## 序列化对象为二进制和Json

### Go

1. 将protobuf信息写入二进制文件
2. 从二进制文件中读取protobuf信息
3. 写入json文件并比较大小

## gRPC 的四种通信方式

1. 类似REST的单请求+单回复
2. 客户端多请求+服务端单回复
3. 客户端单请求+服务端多回复
4. 客户端多请求+服务端多回复

## gRPC 反射

<a href="https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md">grpc反射</a>
gRPC 服务器反射提供有关服务器上可公开访问的 gRPC 服务的信息，并帮助客户端在运行时构造 RPC 请求和响应，而无需预编译的服务信息。
它由 gRPC CLI 使用，可用于自省服务器原型和发送/接收测试 RPC。

### evans

一个grpc客户端,在服务端开启反射并运行时，通过`evans -r repl -p 端口`进入shell，
通过`show package`查看反射的包信息，通过`package 包名`选择不同的包,通过`show service`查看反射的服务信息,通过`service 服务`选择服务，通过`call CreateLaptop`
调用服务，中途使用`ctrl+D`取消重复字段的输入...
<a href="https://github.com/ktr0731/evans">evans</a>

## gRPC 拦截器

类似于中间件，可以在服务端和客户端之间添加的额外功能，服务器端拦截器是gRPC服务器在调用实际RPC方法前将调用的函数，可以用于日志记录，跟踪，限流，身份验证，限流等。
客户端拦截器是gRPC客户端在调用实际RPC方法前将调用的函数.

服务器端拦截器将采用JWT来验证，客户端拦截器将添加JWT到请求。

## SSL/TLS

### 详解

TLS 是传输层安全协议，它用于实现客户端和服务器之间的加密通信。SSL是TLS的前身。

![image-20220504204623342](https://gitee.com/ORaja/picture/raw/master/img/image-20220504204623342.png)

TLS在网络（HTTPS = HTTP+TLS），邮件（SMTPS = SMTP+TLS），文件传输（FTPS = FTP+TLS）等中使用。

作用：

1. 身份验证 证明访问的网站不是伪造的

   将服务器公钥放入到**数字证书**中，解决了冒充的风险

2. 信息加密 交互信息无法被窃取

   通过**混合加密**的方式可以保证信息的**机密性**，解决了窃听的风险

   HTTPS 采用的是**对称加密**和**非对称加密**结合的「混合加密」方式：

   - 在通信建立前采用**非对称加密**的方式交换「会话秘钥」，后续就不再使用非对称加密。
   - 在通信过程中全部使用**对称加密**的「会话秘钥」的方式加密明文数据。

3. 校验机制 无法篡改通信内容

   **摘要算法**用来实现**完整性**，能够为数据生成独一无二的「指纹」，用于校验数据的完整性，解决了篡改的风险。

一般采用ECDHE密钥协商算法生成会话密钥

1. TLS第一次握手

   客户端首先会发一个「**Client Hello**」消息，消息里面有客户端使用的 TLS 版本号、支持的密码套件列表，以及生成的**随机数（\*Client Random\*）**。

2. TLS第二次握手

   服务端收到客户端的「打招呼」返回「**Server Hello**」消息，消息面有服务器确认的 TLS 版本号，也给出了一个**随机数（\*Server Random\*）**，然后从客户端的密码套件列表选择了一个合适的密码套件。接着，服务端为了证明自己的身份，发送「**Certificate**」消息，会把证书也发给客户端。

   因为服务端选择了 ECDHE 密钥协商算法，所以会在发送完证书后，发送「**Server Key Exchange**」消息。

   - 选择了**椭圆曲线**，选好了椭圆曲线相当于椭圆曲线基点 G 也定好了，这些都会公开给客户端；
   - 生成随机数作为服务端椭圆曲线的私钥，保留到本地；
   - 根据基点 G 和私钥计算出**服务端的椭圆曲线公钥**，这个会公开给客户端。

   为了保证这个椭圆曲线的公钥不被第三方篡改，服务端会用 RSA 签名算法给服务端的椭圆曲线公钥做个签名。

3. TLS第三次握手

   客户端收到了服务端的证书后，校验证书是否合法。

   客户端会生成一个随机数作为客户端椭圆曲线的私钥，然后再根据服务端前面给的信息，生成**客户端的椭圆曲线公钥**，然后用「**Client Key Exchange**」消息发给服务端

   **最终的会话密钥，就是用「客户端随机数 + 服务端随机数 + x（ECDHE 算法算出的共享密钥） 」三个材料生成的**。

   算好会话密钥后，客户端会发一个「**Change Cipher Spec**」消息，告诉服务端后续改用对称算法加密通信。

   接着，客户端会发「**Encrypted Handshake Message**」消息，把之前发送的数据做一个摘要，再用对称密钥加密一下，让服务端做个验证，验证下本次生成的对称密钥是否可以正常使用。

4. TLS第四次握手

   最后，服务端也会有一个同样的操作，发「**Change Cipher Spec**」和「**Encrypted Handshake Message**」消息，如果双方都验证加密和解密没问题，那么握手正式完成。于是，就可以正常收发加密的 HTTP 请求和响应了。

5. INSECURE 无安全验证

6. SERVER-SIDE TLS 服务端证书

   服务端采用TLS加密数据，传递证书给客户端，客户端通过CA进行校验

7. MUTUAL SSL 客户端与服务端证书

   双向进行加密并校验

## nginx 负载均衡

### 服务端

客户端发送请求到代理服务器，代理服务器负责负载均衡

便于部署，且适用于面向不确定使用者的环境下。

但是会增加一个跳点，增加延迟。

```nginx
worker_processes  1;

error_log  /var/log/nginx/error.log;

events {
    worker_connections  1024;
}


http {
    access_log  /var/log/nginx/access.log;

    # 上游
    upstream laptop_services {
        server 0.0.0.0:50051;
        server 0.0.0.0:50052;
    }

    server {
        listen       8080 http2;

        location / {
            grpc_pass grpc://laptop_services;
        }
    }
}
```

一般部署情况下，grpc服务器位于安全的环境下，所以只需要让nginx服务器开启SSL/TLS加密。即将服务器私钥，服务器证书以及签署客户端证书的CA证书提供给nginx

```nginx
worker_processes  1;

error_log  /var/log/nginx/error.log;

events {
    worker_connections  1024;
}


http {
    access_log  /var/log/nginx/access.log;

    # 上游
    upstream laptop_services {
        server 0.0.0.0:50051;
        server 0.0.0.0:50052;
    }

    server {
        listen       8080 ssl http2;
        # 服务器证书和密钥
        ssl_certificate         cert/server-cert.pem;
        ssl_certificate_key     cert/server-key.pem;

        # 签署客户端证书的CA证书
        ssl_client_certificate  cert/ca-cert.pem;
        # 开启客户端证书验证
        ssl_verify_client on;

        # grpcs 开启服务端TLS
        location / {
            grpc_pass grpcs://laptop_services;
        }
    }
}
```

但是如果真的需要开启双向TLS，即nginx和grpc服务器之间的双向TLS 则需要将nginx证书传递给grpc服务器

```nginx
worker_processes  1;

error_log  /var/log/nginx/error.log;

events {
    worker_connections  1024;
}


http {
    access_log  /var/log/nginx/access.log;

    # 上游
    upstream laptop_services {
        server 0.0.0.0:50051;
        server 0.0.0.0:50052;
    }

    server {
        listen       8080 ssl http2;
        # 服务器证书和密钥
        ssl_certificate         cert/server-cert.pem;
        ssl_certificate_key     cert/server-key.pem;

        # 签署客户端证书的CA证书
        ssl_client_certificate  cert/ca-cert.pem;
        # 开启客户端证书验证
        ssl_verify_client on;

        # grpcs 开启nginx服务端TLS
        location / {
            grpc_pass grpcs://laptop_services;

            # 开启nginx TLS (可以为nginx生成指定证书) 向服务端发送TLS证书
            grpc_ssl_certificate cert/server-cert.pem;
            grpc_ssl_certificate_key cert/server-key.pem;
        }
    }
}
```

其次实现业务分离

```nginx
worker_processes  1;

error_log  /var/log/nginx/error.log;

events {
    worker_connections  1024;
}


http {
    access_log  /var/log/nginx/access.log;

    # auth上游
    upstream auth_services {
        server 0.0.0.0:50051;
    }
    # laptop上游
    upstream laptop_services {
        server 0.0.0.0:50052;
    }

    server {
        listen       8080 ssl http2;
        # 服务器证书和密钥
        ssl_certificate         cert/server-cert.pem;
        ssl_certificate_key     cert/server-key.pem;

        # 签署客户端证书的CA证书
        ssl_client_certificate  cert/ca-cert.pem;
        # 开启客户端证书验证
        ssl_verify_client on;

        # grpcs 开启nginx服务端TLS
        # 转发auth
        location /rpc.proto.AuthService {
            grpc_pass grpcs://auth_services;

            # 开启nginx TLS (可以为nginx生成指定证书) 向服务端发送TLS证书
            grpc_ssl_certificate cert/server-cert.pem;
            grpc_ssl_certificate_key cert/server-key.pem;
        }
        # 转发laptop
        location /rpc.proto.LaptopService {
            grpc_pass grpcs://laptop_services;

            # 开启nginx TLS (可以为nginx生成指定证书) 向服务端发送TLS证书
            grpc_ssl_certificate cert/server-cert.pem;
            grpc_ssl_certificate_key cert/server-key.pem;
        }
    }
}
```

### 客户端

客户端为每个RPC选择不同的后端服务器，通过服务注册来注册后端服务器，客户端访问服务注册来获取服务器地址。

延迟低，但是实现复杂且适用于安全的场景下。

## grpc 网关

gRPC网关可以通过protobuf服务定义生成代理服务器，然后将REST请求翻译为grpc请求

### 安装插件

```bash
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
go get google.golang.org/protobuf/cmd/protoc-gen-go

# 增加google.api.http 引入第三方protobuf包
cp -r $GOPATH/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0/third_party/googleapis/google ./proto
```

```protobuf
//增加
import "google/api/annotations.proto";
//修改
service AuthService{
  rpc Login(LoginRequest) returns (LoginResponse){
    option (google.api.http) = {
      post: "/v1/auth/login"
      body: "*"
    };
  }
}
```

之后生成grpc网管和swagger文件

```bash
protoc --proto_path=proto --go_out=plugins=grpc:pb proto/*.proto --grpc-gateway_out=:pb --swagger_out=:swagger
# --grpc-gateway_out=:pb 指定网关生成路径
# --swagger_out=:swagger swaager文件生成路径
```

1. 进程间RPC转换

   不需要运行单独的gRPC服务器，但是目前只支持一元RPC

   修改main文件 增加启动REST的方式

   ```go
   type serverConfig struct {
   	laptopServer *service.LaptopServer
   	authServer   *service.AuthServer
   	enableTLS    bool
   	listener     net.Listener
   	maker        token.Maker
   }
   
   func runRESTServer(config *serverConfig) error {
   	mux := runtime.NewServeMux()
   	ctx, cancel := context.WithCancel(context.Background())
   	defer cancel()
   	//gRPC到REST的进程间转换
   	if err := pb.RegisterAuthServiceHandlerServer(ctx, mux, config.authServer); err != nil {
   		return err
   	}
   	if err := pb.RegisterLaptopServiceHandlerServer(ctx, mux, config.laptopServer); err != nil {
   		return err
   	}
   	log.Println("server port:", config.listener.Addr().String(), " tls=", config.enableTLS)
   	if config.enableTLS {
   		return http.ServeTLS(config.listener, mux, ServerCert, ServerKey)
   	}
   	return http.Serve(config.listener, mux)
   }
   func runGRPCServer(config *serverConfig) error {
   	//初始化拦截器
   	interceptor := service.NewAuthInterceptor(config.maker, accessibleRoles())
   	serviceOptions := []grpc.ServerOption{
   		//安装一个一元拦截器
   		grpc.UnaryInterceptor(interceptor.Unary()),
   		//安装一个流拦截器
   		grpc.StreamInterceptor(interceptor.Stream()),
   	}
   	//添加TLS凭证
   	if config.enableTLS {
   		tlsCredentials, err := loadTLSCredentials()
   		if err != nil {
   			return fmt.Errorf("can't load TLS credentials,error: %w", err)
   		}
   		serviceOptions = append(serviceOptions, grpc.Creds(tlsCredentials))
   	}
   	//配置gRPC服务器
   	grpcServer := grpc.NewServer(
   		serviceOptions...,
   	)
   	reflection.Register(grpcServer)                                 //注册反射服务
   	pb.RegisterLaptopServiceServer(grpcServer, config.laptopServer) //注册laptop服务
   	pb.RegisterAuthServiceServer(grpcServer, config.authServer)     //注册auth服务
   	log.Println("server port:", config.listener.Addr().String(), " tls=", config.enableTLS)
   	//启动gRPC服务
   	return grpcServer.Serve(config.listener)
   }
   
   func main() {
   	portPtr := flag.Int("port", 8080, "server port")
   	enableTLSPtr := flag.Bool("tls", false, "enable tls") //是否开启TLS
   	serverType := flag.String("type", "grpc", "type of server(grpc/rest)")
   	flag.Parse()
   	//初始化持久层
   	userStore := service.NewInMemoryUserStoreStore()
   	maker, err := token.NewPasetoMaker([]byte(Secret))
   	if err != nil {
   		log.Fatalln("create maker error:", err)
   	}
   	//初始化用户存储
   	if err := seedUsers(userStore); err != nil {
   		log.Fatalln("cannot seed user store:", err)
   	}
   	laptopStore := service.NewInMemoryLaptopStore()
   	imageStore := service.NewDiskImageStore("img", ImageMaxSize)
   	scoreStore := service.NewInMemoryRateStoreStore()
   
   	laptopServer := service.NewLaptopServer(laptopStore, imageStore, scoreStore)
   	authServer := service.NewAuthServer(userStore, maker)
   	addr := fmt.Sprintf("0.0.0.0:%d", *portPtr)
   	listener, err := net.Listen("tcp", addr)
   	if err != nil {
   		log.Fatalln("listener error: ", err)
   	}
   	config := &serverConfig{
   		laptopServer: laptopServer,
   		authServer:   authServer,
   		enableTLS:    *enableTLSPtr,
   		listener:     listener,
   		maker:        maker,
   	}
   	if *serverType == "grpc" {
   		err = runGRPCServer(config)
   	} else {
   		err = runRESTServer(config)
   	}
   }
   ```

   2, 可以使用grpc网关支持REST流式RPC

   需要同时开启gRPC服务器和REST服务器，REST服务器会将接收到的请求发送到gRPC服务器并返回结果

   ```go
   type serverConfig struct {
   	laptopServer *service.LaptopServer
   	authServer   *service.AuthServer
   	enableTLS    bool
   	listener     net.Listener
   	maker        token.Maker
   	grpcEndpoint string
   }
   
   func runRESTServer(config *serverConfig) error {
   	mux := runtime.NewServeMux()
   	ctx, cancel := context.WithCancel(context.Background())
   	defer cancel()
   	//设置拨号选项
   	dialOpts := []grpc.DialOption{grpc.WithInsecure()}
   	// gRPC到REST的进程间转换 pb.RegisterAuthServiceHandlerServer
   	// gRPC网关 RegisterAuthServiceHandlerFromEndpoint
   	if err := pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, config.grpcEndpoint, dialOpts); err != nil {
   		return err
   	}
   	if err := pb.RegisterLaptopServiceHandlerFromEndpoint(ctx, mux, config.grpcEndpoint, dialOpts); err != nil {
   		return err
   	}
   	log.Println("server port:", config.listener.Addr().String(), " tls=", config.enableTLS, "grpcEndpoint=", config.grpcEndpoint)
   	if config.enableTLS {
   		return http.ServeTLS(config.listener, mux, ServerCert, ServerKey)
   	}
   	return http.Serve(config.listener, mux)
   }
   ```

   

