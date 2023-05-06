
# 用一元拦截器实现认证

## 服务端

```

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

	// 取出uid
	uidTmp, ok := md["uid"]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "uid不存在")
	}
	_ = uidTmp[0]
	//验证
	j := util.NewJWT()
	parseToken, err := j.ParseToken(token)
	if err != nil {
		fmt.Println(err)
		return nil, status.Error(codes.InvalidArgument, "token验证失败")
	}
	fmt.Println("parseToken: ", parseToken.Username)

	return nil, nil
}

```

```go
	// 创建一个gRPC服务器实例。
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			handler.ServerInterceptorCheckToken(),
			//handler.UnaryServerInterceptor()
		),
		//grpc.StreamInterceptor(handler.StreamServerInterceptor()),
	)
```

## 客户端


```go


// Token token认证
type Token struct {
	Uid   string
	Token string
}

// GetRequestMetadata 获取当前请求认证所需的元数据（metadata）
func (t *Token) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	// 设置一个种子
	rand.Seed(time.Now().UnixNano())
	// Intn返回一个取值范围在[0,n)的伪随机int值
	num := rand.Intn(100) + 1 // 随机1-100
	rangeSeed := strconv.Itoa(num)
	log.Println("GetRequestMetadata 每次访问服务端方法都会被调用 添加自定义认证", rangeSeed)

	return map[string]string{"uid": t.Uid, "token": t.Token, "range_seed": rangeSeed}, nil
}

// RequireTransportSecurity 是否需要基于 TLS 认证进行安全传输,返回false不进行TLS验证
func (t *Token) RequireTransportSecurity() bool {
	return true
}

```
 添加认证
	`grpc.WithPerRPCCredentials(&token)`
```go
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
			handler.UnaryClientInterceptor(),
			//handler.UnaryClientInterceptorTwo()
		),
		//流式拦截器
		//grpc.WithStreamInterceptor(handler.StreamClientInterceptor()),
	)
```

## jwt 生成


```go
package util

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type JWT struct {
	SigningKey []byte
}

// CustomClaims  structure
type CustomClaims struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}

type BaseClaims struct {
	ID       uint
	Username string
}

var (
	TokenExpired     = errors.New("token is expired")
	TokenNotValidYet = errors.New("token not active yet")
	TokenMalformed   = errors.New("that's not even a token")
	TokenInvalid     = errors.New("couldn't handle this token")
)

func NewJWT() *JWT {
	return &JWT{
		[]byte("12312dsdsdfdfbndassa"),
	}
}

func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	claims := CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: int64(1 / time.Second), // 缓冲时间1天 缓冲时间内会获得新的token刷新令牌 此时一个用户会存在两个有效令牌 但是前端只留一个 另一个会丢失
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"哈哈"},                                 // 受众
			NotBefore: jwt.NewNumericDate(time.Now()),                         // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 过期时间 7天  配置文件
			Issuer:    "天下",                                                   // 签名的发行者
		},
	}
	return claims
}

// CreateToken 创建一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken 解析 token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid

	} else {
		return nil, TokenInvalid
	}
}

```
