package config

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("2c62f6f19b67f8e2e57826a0470842094e46581981b99561b3dc10c0484e5ea54d75e5981d8e482afb56f123f924f45e4a9a6765eeb6267a2bd7fd49b99d65e367ef6d4d704c5e819f6ad15a0f4d44ceffd6ca5cc3d27a69c89774b7b1a70d654abe74855aff918a4eeb449a07e10e7875dc0ee45acefa3612bc06265823a648dd4947e57a35eff041dcd90252bfbce9d5021ef15a0157a10535012743393eaba3347a6d836844e1cda26062689cb9fbec5dc6f308249b39a5c96e3ac522d2f6681ab3157b51e7a24980672ecf558028e3db39fd694da43c23aaf0cf967f340e5f82de7abb92ca1146ef9cd936b1d8a8b2aa8a17545ccfa226f2309959e97873") // In production, use environment variable

func GenerateToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hour expiry
	})

	return token.SignedString(JWTSecret)
}
