package util

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
)

// GenerateToken 生成token
//
//	expireAt: 过期时间，时间戳，秒
func (uc *UserClaim) GenerateToken(expireAt int64) (string, error) {
	uc.StandardClaims = jwt.StandardClaims{
		ExpiresAt: expireAt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
	tokenString, err := token.SignedString([]byte(JwtKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AnalyzeToken 解析token
func AnalyzeToken(token string) (*UserClaim, error) {
	uc := new(UserClaim)
	claims, err := jwt.ParseWithClaims(token, uc, func(token *jwt.Token) (interface{}, error) {
		return []byte(JwtKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return uc, errors.New("token is invalid")
	}
	return uc, err
}
