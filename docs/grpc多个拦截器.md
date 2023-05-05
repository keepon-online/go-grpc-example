## 实现多个拦截器

gRPC框架中只能为每个服务一起配置一元和流拦截器，，gRPC 会根据不同方法选择对应类型的拦截器执行，因此所有的工作只能在一个函数中完成。

```go

func UnaryClientInterceptorTwo() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Printf("第二个拦截器 \n")
		return nil
	}
}

```

// 按照顺序依次执行截取器
```
	addr := "192.168.2.166:8080"
	// 使用 grpc.Dial 创建一个到指定地址的 gRPC 连接。
	// 此处使用不安全的证书来实现 SSL/TLS 连接
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//普通拦截器
		grpc.WithChainUnaryInterceptor(
			handler.UnaryClientInterceptor(),
			handler.UnaryClientInterceptorTwo()),
		//流式拦截器
		grpc.WithStreamInterceptor(handler.StreamClientInterceptor()),
	)

```
