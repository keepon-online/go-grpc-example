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
