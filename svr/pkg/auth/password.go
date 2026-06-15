package auth

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// pepper 获取密码哈希使用的环境级加密因子。
func pepper() string {
	if p := os.Getenv("PASSWORD_PEPPER"); p != "" {
		return p
	}
	return "wiki-default-pepper"
}

// Hash 对密码进行 bcrypt 哈希，自动加 pepper 和随机盐值。
func Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password+pepper()), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("密码哈希失败: %w", err)
	}
	return string(hash), nil
}

// Verify 验证密码是否匹配哈希值。
func Verify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+pepper()))
	return err == nil
}
