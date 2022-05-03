package token

import (
	"errors"
	"time"
)

var (
	ErrSecretLen = errors.New("密钥长度不正确")
	ErrTimeOut   = errors.New("超时")
)

type Maker interface {
	// CreateToken 生成Token
	CreateToken(username, role string, expireDate time.Duration) (string, *Payload, error)
	// VerifyToken 解析Token
	VerifyToken(token string) (*Payload, error)
}
