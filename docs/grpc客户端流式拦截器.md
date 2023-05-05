## 

```go


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
```
## 使用

```go
conn, err := grpc.Dial(addr,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    //普通拦截器
    grpc.WithUnaryInterceptor(handler.UnaryClientInterceptor()),
    //流式拦截器
    grpc.WithStreamInterceptor(handler.StreamClientInterceptor()))
```

