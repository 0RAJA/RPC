package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User 存储用户信息
type User struct {
	Username     string
	HashPassword []byte
	Role         string
}

func NewUser(username, password, role string) (*User, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate:%v", err)
	}
	return &User{Username: username, HashPassword: hashPassword, Role: role}, nil
}

func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(user.HashPassword, []byte(password))
	return err == nil
}

func (user *User) Clone() *User {
	return &User{
		Username:     user.Username,
		HashPassword: user.HashPassword,
		Role:         user.Role,
	}
}
