# 使用服务器身份验证 SSL/TLS
gRPC 内置支持 SSL/TLS，可以通过 SSL/TLS 证书建立安全连接，对传输的数据进行加密处理。

## 生成私钥
执行下面的命令生成私钥文件——`server.key`。
```
openssl ecparam -genkey -name secp384r1 -out server.key
```
## 生成自签名的证书

为了在证书中添加SANs信息，我们将下面自定义配置保存到server.cnf文件中
```
[ req ]
default_bits       = 4096
default_md		= sha256
distinguished_name = req_distinguished_name
req_extensions     = req_ext

[ req_distinguished_name ]
countryName                 = Country Name (2 letter code)
countryName_default         = CN
stateOrProvinceName         = State or Province Name (full name)
stateOrProvinceName_default = BEIJING
localityName                = Locality Name (eg, city)
localityName_default        = BEIJING
organizationName            = Organization Name (eg, company)
organizationName_default    = DEV
commonName                  = Common Name (e.g. server FQDN or YOUR name)
commonName_max              = 64
commonName_default          = keepon.online

[ req_ext ]
subjectAltName = @alt_names

[alt_names]
DNS.1   = localhost
DNS.2   = keepon.online
IP      = 127.0.0.1
IP      = 192.168.2.166




```
执行下面的命令生成自签名证书——`server.crt`
```
openssl req -nodes -new -x509 -sha256 -days 3650 -config server.cnf -extensions 'req_ext' -key server.key -out server.crt
```

## 服务端

```
	creds, _ := credentials.NewServerTLSFromFile("conf/server.crt", "conf/server.key")

	// 创建一个gRPC服务器实例。
	s := grpc.NewServer(
		grpc.Creds(creds))
		
```

## 客户端

```
creds, _ := credentials.NewClientTLSFromFile("conf/server.crt", "")
conn, _ := grpc.Dial("192.168.2.166:8080", grpc.WithTransportCredentials(creds))
client := pb.NewGreeterClient(conn)

```
