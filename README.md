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

### gRPC 拦截器

类似于中间件，可以在服务端和客户端之间添加的额外功能，服务器端拦截器是gRPC服务器在调用实际RPC方法前将调用的函数，可以用于日志记录，跟踪，限流，身份验证，限流等。
客户端拦截器是gRPC客户端在调用实际RPC方法前将调用的函数.

服务器端拦截器将采用JWT来验证，客户端拦截器将添加JWT到请求。
