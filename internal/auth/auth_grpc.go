package auth

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type AuthManager struct{}

func NewAuthManager() *AuthManager {
	return &AuthManager{}
}

func (am *AuthManager) IsLogined(ctx context.Context, in *AuthCookie) (*BoolResult, error) {
	_, err := am.GetUserId(ctx, in)
	return &BoolResult{BoolResult: err == nil}, nil
}

func (am *AuthManager) GetUserId(ctx context.Context, in *AuthCookie) (*IdResult, error) {
	token, err := jwt.Parse(in.Data, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(in.Secret), nil
	})
	if err != nil {
		return &IdResult{Id: 0}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		_, ok := claims["expires"]
		if !ok {
			return &IdResult{Id: 0}, fmt.Errorf("no expires")
		}
		_, ok = claims["userId"]
		if !ok {
			return &IdResult{Id: 0}, fmt.Errorf("no userId")
		}
		expires, ok := claims["expires"].(float64)
		if time.Now().Unix() > int64(expires) {
			return &IdResult{Id: 0}, fmt.Errorf("expired")
		}
		userId, ok := claims["userId"].(float64)
		if !ok {
			return &IdResult{Id: 0}, fmt.Errorf("wrong userId type")
		}
		return &IdResult{Id: int64(userId)}, nil
	}
	return &IdResult{Id: 0}, fmt.Errorf("token invalid")
}
