
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

### 服务端实现流接口

```go
// LotsOfReplies 服务端返回流
func (s HelloServer) LotsOfReplies(in *hello.HelloRequest, stream hello.HelloService_LotsOfRepliesServer) error {
	words := []string{
		"你好",
		"hello",
		"こんにちは",
		"안녕하세요",
	}
	for _, word := range words {
		data := &hello.HelloResponse{
			Name: in.Name + word,
		}
		// 使用Send方法返回多个数据
		if err := stream.Send(data); err != nil {
			return err
		}
	}
	return nil
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
	//接收服务端流
	runLotsOfReplies(client, &helloRequest)
	fmt.Println(result)
}

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
		log.Printf("got reply: %q\n", res.GetName())
	}
}

```

```
2023/05/05 15:24:01 got reply: "鲁迪你好"
2023/05/05 15:24:01 got reply: "鲁迪hello"
2023/05/05 15:24:01 got reply: "鲁迪こんにちは"
2023/05/05 15:24:01 got reply: "鲁迪안녕하세요"



```
