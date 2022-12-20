package middleware

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"neko-question-box-be/internal/config"
	"neko-question-box-be/internal/database/types"
	"net/http"
	"time"
)

//var jwtKey = []byte("")

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var (
	ErrTokenExpired          = errors.New("令牌已过期")
	ErrTokenSignatureInvalid = errors.New("签名无效")
)

func IssueToken(userProfile types.User) (string, error) {
	// set the expiration time of the token
	expirationTime := time.Now().Add(3 * time.Minute)
	claims := &Claims{
		Username: userProfile.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Issuer:    "NekoQuestionBox",
		},
	}

	// 使用用于签名的算法和令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 创建JWT字符串
	tokenString, err := token.SignedString([]byte(config.Conf.JwtKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

//解析 Token ，提取并返回 Claims 给调用这个函数的函数
func ParseToken(ctx *gin.Context) (*jwt.Token, *Claims, error) {
	cookie, err := ctx.Request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			ctx.String(http.StatusUnauthorized, "you are not login.")
			return nil, nil, err
		}
		return nil, nil, err
	}
	MyClaims := &Claims{}

	// 解析JWT字符串并将结果存储在`claims`中。
	tkn, err := jwt.ParseWithClaims(cookie.Value, MyClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Conf.JwtKey), nil
	})

	if err != nil {
		//如果 JWT Token 解析到了【已过期】错误
		ValidationError, ok := err.(*jwt.ValidationError)
		if ok && ValidationError != nil {
			if ValidationError.Errors&jwt.ValidationErrorExpired != 0 {
				ctx.String(http.StatusUnauthorized, "Token Expired!")
				ctx.Abort()
				return nil, nil, ErrTokenExpired
			}
			if errors.Is(ValidationError.Inner, jwt.ErrSignatureInvalid) {
				ctx.String(http.StatusUnauthorized, "Token Signature Invalid!")
				ctx.Abort()
				return nil, nil, ErrTokenSignatureInvalid
			}
		}
	}
	return tkn, MyClaims, err
}

// 中间件，用于检查请求是否携带了合法的令牌。
func JwtToken(ctx *gin.Context) {
	fmt.Println("Authenticating users...")

	// 初始化`Claims`实例
	_, MyClaims, err := ParseToken(ctx)
	if err != nil {
		ctx.Abort()
		return
	}
	ctx.Set("username", MyClaims)
	ctx.Next()
}
