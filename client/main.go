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
