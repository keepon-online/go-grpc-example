package handler

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/keepon-online/go-grpc-example/gen/hello"
	"github.com/keepon-online/go-grpc-example/util"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"log"
	"strings"
	"time"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 预处理(pre-processing)
		start := time.Now()
		// 从传入上下文获取元数据
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, fmt.Errorf("couldn't parse incoming context metadata")
		}

		// 检索客户端操作系统，如果它不存在，则此值为空
		os := md.Get("client-os")
		// 获取客户端IP地址
		ip, err := getClientIP(ctx)
		if err != nil {
			return nil, err
		}

		// RPC 方法真正执行的逻辑
		// 调用RPC方法(invoking RPC method)
		m, err := handler(ctx, req)
		end := time.Now()
		// 记录请求参数 耗时 错误信息等数据
		// 后处理(post-processing)
		log.Printf("RPC: %s,client-OS: '%v' and IP: '%v' req:%v start time: %s, end time: %s, err: %v", info.FullMethod, os, ip, req, start.Format(time.RFC3339), end.Format(time.RFC3339), err)
		return m, err
	}
}

// GetClientIP检查上下文以检索客户机的ip地址
func getClientIP(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("couldn't parse client IP address")
	}
	return p.Addr.String(), nil
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream,
		info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := newStreamServer(ss)
		return handler(srv, wrapper)
	}
}

// 嵌入式EdgeServerStream允许我们访问RecvMsg函数
type streamServer struct {
	grpc.ServerStream
}

func newStreamServer(s grpc.ServerStream) grpc.ServerStream {
	return &streamServer{s}
}

// RecvMsg从流中接收消息
func (e *streamServer) RecvMsg(m interface{}) error {
	// 在这里，我们可以对接收到的消息执行额外的逻辑，例如
	// 验证
	log.Printf("Receive a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	if err := e.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	return nil
}

// RecvMsg从流中接收消息
func (e *streamServer) SendMsg(m interface{}) error {
	// 在这里，我们可以对接收到的消息执行额外的逻辑，例如
	// 验证
	log.Printf("Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	if err := e.ServerStream.SendMsg(m); err != nil {
		return err
	}
	return nil
}

// ServerInterceptorCheckToken 用一元拦截器实现认证
func ServerInterceptorCheckToken() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 验证token
		_, err = checkToken(ctx)
		if err != nil {
			fmt.Println("Interceptor 拦截器内token认证失败")
			return nil, err
		}
		fmt.Println("Interceptor 拦截器内token认证成功")
		return handler(ctx, req)
	}
}

// 验证
func checkToken(ctx context.Context) (*hello.HelloResponse, error) {
	// 取出元数据
	md, b := metadata.FromIncomingContext(ctx)
	if !b {
		return nil, status.Error(codes.InvalidArgument, "token信息不存在")
	}

	var token, _ string
	// 取出token
	tokenInfo, ok := md["token"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "token不存在")
	}

	token = tokenInfo[0]

	//验证
	j := util.NewJWT()
	parseToken, err := j.ParseToken(token)
	if err != nil {
		st := status.New(codes.InvalidArgument, "token校验失败")
		ds, err := st.WithDetails(
			&errdetails.QuotaFailure{
				Violations: []*errdetails.QuotaFailure_Violation{{
					Subject:     fmt.Sprintf("token失效"),
					Description: "请使用新的token",
				}},
			},
		)
		if err != nil {
			return nil, st.Err()
		}
		return nil, ds.Err()
	}
	fmt.Println("parseToken: ", parseToken.Username)
	return nil, nil
}

// AuthenticateInterceptor 定义一个认证拦截器，将token添加到gRPC元数据中进行身份验证
func AuthenticateInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 将token添加到gRPC元数据中
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	ctx = metadata.NewIncomingContext(ctx, md)
	// 调用下一个处理程序
	return handler(ctx, req)
}

// CustomHeaderMatcher 定义一个HTTP请求处理程序，将token从自定义头中提取出来，并将其添加到gRPC元数据中进行身份验证
func CustomHeaderMatcher(key string) (string, bool) {
	if strings.ToLower(key) == "authorization" {
		return "token", true
	}
	return runtime.DefaultHeaderMatcher(key)
}
