
# proto 添加

```protobuf

syntax = "proto3";

// 指定proto版本
package hello.v1;

// 指定默认包名

// 指定golang包名
option go_package = "github.com/keepon-online/go-grpc-example;hello";

//定义rpc服务
service HelloService {
    // 定义函数
    rpc SayHello (HelloRequest) returns (HelloResponse) {}

    // 服务端返回流式数据
    rpc LotsOfReplies (HelloRequest) returns (stream HelloResponse);

    // 客户端发送流式数据
    rpc LotsOfGreetings(stream HelloRequest) returns (HelloResponse);
}

// HelloRequest 请求内容
message HelloRequest {
    string name = 1;
    string message = 2;
}

// HelloResponse 响应内容
message HelloResponse {
    string name = 1;
    string message = 2;
}


```
## 重新生成代码

### 服务端接收流接口

```go

// LotsOfGreetings 接收流式数据
func (s HelloServer) LotsOfGreetings(stream hello.HelloService_LotsOfGreetingsServer) error {
	reply := "你好："
	for {
		// 接收客户端发来的流式数据
		res, err := stream.Recv()
		if err == io.EOF {
			// 最终统一回复
			return stream.SendAndClose(&hello.HelloResponse{
				Name: reply,
			})
		}
		if err != nil {
			return err
		}
		reply += res.GetName()
	}
}
```

### 客户端


```go
package main

import (
	"context"
	"fmt"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"time"
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
	helloRequest := hello.HelloRequest{
		Name:    "鲁迪",
		Message: "ok",
	}
	result, err := client.SayHello(context.Background(), &helloRequest)
	fmt.Println(result)
	//接收服务端流
	runLotsOfReplies(client, &helloRequest)
	//向服务端发送流
	runLotsOfGreeting(client)
}

// 接收服务端流
func runLotsOfReplies(c hello.HelloServiceClient, request *hello.HelloRequest) {
	// server端流式RPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := c.LotsOfReplies(ctx, request)
	if err != nil {
		log.Fatalf("c.LotsOfReplies failed, err: %v", err)
	}
	for {
		// 接收服务端返回的流式数据，当收到io.EOF或错误时退出
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("c.LotsOfReplies failed, err: %v", err)
		}
		log.Printf("接收服务端流 reply: %q\n", res.GetName())
	}
}

// 向服务端发送流
func runLotsOfGreeting(c hello.HelloServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 客户端流式RPC
	stream, err := c.LotsOfGreetings(ctx)
	if err != nil {
		log.Fatalf("c.LotsOfGreetings failed, err: %v", err)
	}
	names := []string{"孙悟空", "齐天大圣", "弼马温"}
	for _, name := range names {
		// 发送流式数据
		err := stream.Send(&hello.HelloRequest{
			Name: name,
		})
		if err != nil {
			log.Fatalf("c.LotsOfGreetings stream.Send(%v) failed, err: %v", name, err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("c.LotsOfGreetings failed: %v", err)
	}
	log.Printf("向服务端发送流 reply: %v", res.GetName())
}


```

```
name:"鲁迪"  message:"ok"
2023/05/05 15:57:56 接收服务端流 reply: "鲁迪你好"
2023/05/05 15:57:56 接收服务端流 reply: "鲁迪hello"
2023/05/05 15:57:56 接收服务端流 reply: "鲁迪こんにちは"
2023/05/05 15:57:56 接收服务端流 reply: "鲁迪안녕하세요"
2023/05/05 15:57:56 向服务端发送流 reply: 你好：孙悟空齐天大圣弼马温




```
