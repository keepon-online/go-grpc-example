# GRPC案例

## 必要的环境安装

```shell
 go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

- [GRPC简单实现](docs/grpc简单实现.md)
- [服务端流](docs/grpc服务端流.md)
- [客户端流](docs/grpc客户端流.md)
- [服务端客户端双向流](docs/grpc服务端客户端双向流.md)
- [客户端普通拦截器](docs/grpc客户端普通拦截器.md)
- [客户端流式拦截器](docs/grpc客户端流式拦截器.md)
- [服务端普通拦截器](docs/grpc服务端普通拦截器.md)
- [服务端流式拦截器](docs/grpc服务端流式拦截器.md)
- [多个拦截器](docs/grpc多个拦截器.md)
- [使用服务器身份验证 SSL/TLS](docs/使用服务器身份验证SSL-TLS.md)
- [实现Token认证](docs/实现Token认证.md)
- [gRPC-Gateway](docs/gRPC-Gateway.md)
- [gRPC-错误处理](docs/gRPC-错误处理.md)


## 参考

[gprc使用实践](https://zhuyasen.com/post/grpc.html)
