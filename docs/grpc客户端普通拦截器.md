
## 普通拦截器

```go

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
//这是我们可以使用客户端元数据丰富消息的地方，例如有关客户端运行的硬件或操作系统的一些信息，或者可能启动我们的跟踪流程
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

```
## 使用
````go

	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//普通拦截器
		grpc.WithUnaryInterceptor(handler.UnaryClientInterceptor()),
	)
````
