package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"github.com/keepon-online/go-grpc-example/server/handler"
	"github.com/keepon-online/go-grpc-example/server/service"
	"github.com/keepon-online/go-grpc-example/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"io"
	"log"
	"net"
	"net/http"
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

type GateWayServer struct {
	hello.UnimplementedGatewayServiceServer
}

func (g GateWayServer) SayMessage(ctx context.Context, request *hello.HelloRequest) (*hello.HelloResponse, error) {
	log.Printf("SayMessage %v\n", request)
	return &hello.HelloResponse{
		Message: request.GetMessage(),
		Name:    request.GetName(),
	}, nil
}

func main() {
	// 监听端口
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
	}
	creds, _ := credentials.NewServerTLSFromFile("conf/server.crt", "conf/server.key")

	// 创建一个gRPC服务器实例。
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			handler.ServerInterceptorCheckToken(),
			handler.AuthenticateInterceptor,
			handler.GrpcRecover(),
			//handler.UnaryServerInterceptor()
		),
		//grpc.StreamInterceptor(handler.StreamServerInterceptor()),
	)
	// 将server结构体注册为gRPC服务。
	hello.RegisterHelloServiceServer(s, &HelloServer{})
	hello.RegisterGatewayServiceServer(s, &GateWayServer{})
	hello.RegisterFileServiceServer(s, &service.FileServer{})
	fmt.Println("grpc server running :8080")
	//tlsConfig := util.GetTLSConfig("conf/server.crt", "conf/server.key")
	// NewListener将会创建一个Listener
	// 它接受两个参数，第一个是来自内部Listener的监听器，第二个参数是tls.Config（必须包含至少一个证书）
	go func() {
		//httpServer(s, ":8081", tlsConfig)
		httpSe()
		log.Printf("----go httpServer---")

	}()
	if err = s.Serve(listen); err != nil {
		log.Printf("ListenAndServe: %v\n", err)
	}

}

func httpServer(grpcServer *grpc.Server, EndPoint string, tlsConfig *tls.Config) *http.Server {
	log.Printf("----httpServer---")
	// 创建 grpc-gateway 关联组件
	// context.Background()返回一个非空的空上下文。
	// 它没有被注销，没有值，没有过期时间。它通常由主函数、初始化和测试使用，并作为传入请求的顶级上下文
	ctx := context.Background()

	// 从客户端的输入证书文件构造TLS凭证
	dcreds, err := credentials.NewClientTLSFromFile("conf/server.crt", "")
	if err != nil {
		log.Printf("Failed to create client TLS credentials %v", err)
	}
	// grpc.WithTransportCredentials 配置一个连接级别的安全凭据(例：TLS、SSL)，返回值为type DialOption
	// grpc.DialOption DialOption选项配置我们如何设置连接（其内部具体由多个的DialOption组成，决定其设置连接的内容）

	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	// 创建HTTP NewServeMux及注册grpc-gateway逻辑
	// runtime.NewServeMux：返回一个新的ServeMux，它的内部映射是空的；
	// ServeMux是grpc-gateway的一个请求多路复用器。它将http请求与模式匹配，并调用相应的处理程序
	gwmux := runtime.NewServeMux()
	// RegisterGatewayServiceHandlerFromEndpoint：注册HelloWorld服务的HTTP Handle到grpc端点
	if err := hello.RegisterGatewayServiceHandlerFromEndpoint(ctx, gwmux, EndPoint, dopts); err != nil {
		log.Printf("Failed to register gw server: %v\n", err)
	}

	// http服务
	// 分配并返回一个新的ServeMux
	mux := http.NewServeMux()
	// 为给定模式注册处理程序
	mux.Handle("/", gwmux)

	return &http.Server{
		Addr:      EndPoint,
		Handler:   util.GrpcHandlerFunc2(grpcServer, mux),
		TLSConfig: tlsConfig,
	}

}

func httpSe() {
	// 2. 启动 HTTP 服务
	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests

	// 从客户端的输入证书文件构造TLS凭证
	dcreds, err := credentials.NewClientTLSFromFile("conf/server.crt", "")
	if err != nil {
		log.Printf("Failed to create client TLS credentials %v", err)
	}
	conn, err := grpc.Dial(
		fmt.Sprintf("192.168.2.166:%d", 8080),
		grpc.WithTransportCredentials(dcreds),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}
	gwmux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(handler.CustomHeaderMatcher))
	// Register Greeter
	err = hello.RegisterGatewayServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8081),
		Handler: gwmux,
	}
	log.Println("Serving gRPC-Gateway on http://0.0.0.0" + fmt.Sprintf(":%d", 8081))
	log.Fatalln(gwServer.ListenAndServe())

}
