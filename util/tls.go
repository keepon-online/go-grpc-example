package util

import (
	"crypto/tls"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"os"
	"strings"
)

// GetTLSConfig 用于处理从证书凭证文件（PEM），最终获取tls.Config作为HTTP2的使用参数
func GetTLSConfig(certPemPath, certKeyPath string) *tls.Config {
	var certKeyPair *tls.Certificate
	cert, _ := os.ReadFile(certPemPath)
	key, _ := os.ReadFile(certKeyPath)
	// 从一对PEM编码的数据中解析公钥/私钥对。成功则返回公钥/私钥对
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		log.Printf("TLS KeyPair err: %v\n\n", err)
	}
	certKeyPair = &pair
	return &tls.Config{
		// tls.Certificate：返回一个或多个证书，实质我们解析PEM调用的X509KeyPair的函数声明
		// 就是func X509KeyPair(certPEMBlock, keyPEMBlock []byte) (Certificate, error)，返回值就是Certificate
		Certificates: []tls.Certificate{*certKeyPair},
		// http2.NextProtoTLS：NextProtoTLS是谈判期间的NPN/ALPN协议，用于HTTP/2的TLS设置
		NextProtos: []string{http2.NextProtoTLS},
	}
}

// GrpcHandlerFunc 将gRPC请求和HTTP请求分别调用不同的handler处理。
func GrpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func GrpcHandlerFunc2(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}
