package util

import "github.com/dgrijalva/jwt-go"

type UserClaim struct {
	Username string `json:"username"` // 用户名
	Level    string `json:"level"`    // 用户等级, root:超级管理员
	jwt.StandardClaims
}
