package handler

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"runtime"
	"time"
)

// UnaryClientInterceptor 普通拦截器实现
// 这是我们可以使用客户端元数据丰富消息的地方，例如有关客户端运行的硬件或操作系统的一些信息，或者可能启动我们的跟踪流程
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 预处理(pre-processing)
		start := time.Now()
		// 获取正在运行程序的操作系统
		cos := runtime.GOOS
		// 将操作系统信息附加到传出请求
		ctx = metadata.AppendToOutgoingContext(ctx, "client-os", cos)

		// 可以看做是当前 RPC 方法，一般在拦截器中调用 invoker 能达到调用 RPC 方法的效果，当然底层也是 gRPC 在处理。
		// 调用RPC方法(invoking RPC method)
		err := invoker(ctx, method, req, reply, cc, opts...)

		// 后处理(post-processing)
		end := time.Now()
		log.Printf("RPC: %s,,client-OS: '%v' req:%v start time: %s, end time: %s, err: %v", method, cos, req, start.Format(time.RFC3339), end.Format(time.RFC3339), err)
		return err
	}
}

func UnaryClientInterceptorTwo() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Printf("第二个拦截器 \n")
		return nil
	}
}

// StreamClientInterceptor 流式拦截器
// 作用：例如，如果我们将 100 个对象的列表传输到服务器，例如文件或视频的块，我们可以在发送每个块之前拦截，并验证校验和等内容是否有效，将元数据添加到帧等。
// 本例中通过结构体嵌入的方式，对 Streamer 进行包装，在 SendMsg 和 RecvMsg 之前打印出具体的值。
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
		method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		log.Printf("opening client streaming to the server method: %v", method)
		// 调用Streamer函数，获得ClientStream
		stream, err := streamer(ctx, desc, cc, method)
		return newStreamClient(stream), err
	}
}

// 嵌入式 streamClient 允许我们访问SendMsg和RecvMsg函数
type streamClient struct {
	grpc.ClientStream
}

// 对ClientStream进行包装
func newStreamClient(c grpc.ClientStream) grpc.ClientStream {
	return &streamClient{c}
}

// RecvMsg从流中接收消息
func (e *streamClient) RecvMsg(m interface{}) error {
	// 在这里，我们可以对接收到的消息执行额外的逻辑，例如
	// 验证
	log.Printf("Receive a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	if err := e.ClientStream.RecvMsg(m); err != nil {
		return err
	}
	return nil
}

// RecvMsg从流中接收消息
func (e *streamClient) SendMsg(m interface{}) error {
	// 在这里，我们可以对接收到的消息执行额外的逻辑，例如
	// 验证
	log.Printf("Send a message (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	if err := e.ClientStream.SendMsg(m); err != nil {
		return err
	}
	return nil
}
