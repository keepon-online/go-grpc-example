package main

import (
	"context"
	"fmt"
	"github.com/keepon-online/go-grpc-example/client/handler"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"github.com/keepon-online/go-grpc-example/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"time"
)

func main() {
	addr := "192.168.2.166:8080"
	// 使用 grpc.Dial 创建一个到指定地址的 gRPC 连接。
	// 此处使用不安全的证书来实现 SSL/TLS 连接
	//构建Token
	token := handler.Token{
		Uid:   "1234",
		Token: token(),
	}
	creds, _ := credentials.NewClientTLSFromFile("conf/server.crt", "")

	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(&token),
		//普通拦截器
		grpc.WithChainUnaryInterceptor(
		//handler.UnaryClientInterceptor(),
		//handler.UnaryClientInterceptorTwo()
		),
		//流式拦截器
		//grpc.WithStreamInterceptor(handler.StreamClientInterceptor()),
	)
	if err != nil {
		log.Fatalf(fmt.Sprintf("grpc connect addr [%s] 连接失败 %s", addr, err))
	}
	defer conn.Close()
	// 初始化客户端
	client := hello.NewHelloServiceClient(conn)
	gatewayServiceClient := hello.NewGatewayServiceClient(conn)
	helloRequest := hello.HelloRequest{
		Name:    "鲁迪",
		Message: "ok",
	}
	sayMessage(gatewayServiceClient)
	result, err := client.SayHello(context.Background(), &helloRequest)
	fmt.Println(result)
	//接收服务端流
	runLotsOfReplies(client, &helloRequest)
	//向服务端发送流
	runLotsOfGreeting(client)
	// 双向流数据
	runBidiHello(client)
}

func token() string {
	j := util.NewJWT()
	claims := j.CreateClaims(util.BaseClaims{
		ID:       1,
		Username: "hello",
	})
	createToken, _ := j.CreateToken(claims)
	fmt.Println(createToken)
	return createToken
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

// 双向流数据
func runBidiHello(c hello.HelloServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// 双向流模式
	stream, err := c.BidiHello(ctx)
	if err != nil {
		log.Fatalf("c.BidiHello failed, err: %v", err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			// 接收服务端返回的响应
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("c.BidiHello stream.Recv() failed, err: %v", err)
			}
			fmt.Printf("双向流数据-接收服务端返回的响应 ：%s\n", in.GetName())
		}
	}()

	names := []string{"孙悟空", "齐天大圣", "弼马温"}
	for _, name := range names {
		// 发送流式数据
		err := stream.Send(&hello.HelloRequest{
			Name: name,
		})
		if err != nil {
			log.Fatalf("双向流数据-客户端 stream.Send(%v) failed, err: %v", name, err)
		}
	}
	stream.CloseSend()
	<-waitc
}

// gateway
func sayMessage(c hello.GatewayServiceClient) {

	message, err := c.SayMessage(context.Background(), &hello.HelloRequest{Name: "test", Message: "收到请求"})
	if err != nil {

		log.Fatalf("sayMessage error %s", err.Error())
		return
	}
	fmt.Printf("get sayMessage rely name%s message %s\n", message.GetName(), message.GetMessage())
}
