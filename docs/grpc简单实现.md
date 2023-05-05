# grpc 简单实现

# proto 文件

```protobuf

syntax = "proto3"; // 指定proto版本
package hello.v1;     // 指定默认包名

// 指定golang包名
option go_package = "github.com/keepon-online/go-grpc-example;hello";

//定义rpc服务
service HelloService {
    // 定义函数
    rpc SayHello (HelloRequest) returns (HelloResponse) {}
}

// HelloRequest 请求内容
message HelloRequest {
    string name = 1;
    string message = 2;
}

// HelloResponse 响应内容
message HelloResponse{
    string name = 1;
    string message = 2;
}

```

## 代码生成

```yaml
version: v1
plugins:
  - plugin: go
    out: ./gen/hello
    opt:
      - paths=source_relative
  - plugin: go-grpc
    out: ./gen/hello
    opt:
      - paths=source_relative

```

```
    buf generate proto
```


## 服务端

```go
package main

import (
	"context"
	"fmt"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"net"
)

// HelloServer HelloServer 实现HelloServiceServer
type HelloServer struct {
	hello.UnimplementedHelloServiceServer
}

func (s HelloServer) SayHello(ctx context.Context, request *hello.HelloRequest) (pd *hello.HelloResponse, err error) {
	fmt.Println("入参：", request.Name, request.Message)
	return &hello.HelloResponse{
		Name:    request.Name,
		Message: request.Message,
	}, nil
}

func main() {
	// 监听端口
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	// 创建一个gRPC服务器实例。
	s := grpc.NewServer()
	server := HelloServer{}
	// 将server结构体注册为gRPC服务。
	hello.RegisterHelloServiceServer(s, &server)
	fmt.Println("grpc server running :8080")
	// 开始处理客户端请求。
	err = s.Serve(listen)
}

```


## 客户端

```go
package main

import (
	"context"
	"fmt"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func main() {
	addr := "192.168.2.166:8080"
	// 使用 grpc.Dial 创建一个到指定地址的 gRPC 连接。
	// 此处使用不安全的证书来实现 SSL/TLS 连接
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf(fmt.Sprintf("grpc connect addr [%s] 连接失败 %s", addr, err))
	}
	defer conn.Close()
	// 初始化客户端
	client := hello.NewHelloServiceClient(conn)
	result, err := client.SayHello(context.Background(), &hello.HelloRequest{
		Name:    "鲁迪",
		Message: "ok",
	})

	fmt.Println(result)
}

```

