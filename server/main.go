package main

import (
	"context"
	"fmt"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"github.com/keepon-online/go-grpc-example/server/handler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"io"
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

// BidiHello 双向流数据
func (s HelloServer) BidiHello(stream hello.HelloService_BidiHelloServer) error {
	for {
		// 接收流式请求
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		reply := in.GetName() // 对收到的数据做些处理

		// 返回流式响应
		if err := stream.Send(&hello.HelloResponse{Name: reply}); err != nil {
			return err
		}
	}
}

func main() {
	// 监听端口
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}

	// 创建一个gRPC服务器实例。
	s := grpc.NewServer(
		grpc.UnaryInterceptor(handler.UnaryServerInterceptor()),
		grpc.StreamInterceptor(handler.StreamServerInterceptor()),
	)
	server := HelloServer{}
	// 将server结构体注册为gRPC服务。
	hello.RegisterHelloServiceServer(s, &server)
	fmt.Println("grpc server running :8080")
	// 开始处理客户端请求。
	err = s.Serve(listen)
}
