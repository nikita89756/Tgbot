package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Middleware struct{
	Token string
	Logger *zap.Logger
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		token = extractTokenFromHeader(token)
		m.Logger.Info("token", zap.String("token", token))

		if token == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing"})
			ctx.Abort()
			return
		}

		// Парсим токен
		id, err := ParseToken(m.Token, token)
		if err != nil {
			m.Logger.Error("Failed to parse token", zap.Error(err))
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}
		ctx.Set("userId", id)
	}
}
type tokenClaims struct{
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

func extractTokenFromHeader(header string) string {
	if len(header) >7 && strings.HasPrefix(header, "Bearer "){ 
		return header[strings.Index(header, "Bearer ")+len("Bearer "):]
	}
	return ""
}

func GenerateToken(username,password string , id int,jwtkey string , tokenTTL time.Duration)(string,error){
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,&tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		id,
	})
	return token.SignedString([]byte(jwtkey))
}

func ParseToken(jwtToken,token string)(int , error){
	t,err:= jwt.ParseWithClaims(token,&tokenClaims{},func(t *jwt.Token)(interface{},error){
		if _,ok := t.Method.(*jwt.SigningMethodHMAC);!ok{
			return nil , errors.New("invalid signing method")
		}
		return []byte(jwtToken),nil
	})

	if err!=nil{
		return 0 , err
	}

	claims , ok := t.Claims.(*tokenClaims)
	if !ok{
		return 0 , errors.New("bad token")
	}
	
	return claims.UserId , nil
}