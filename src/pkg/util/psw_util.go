package util

import "golang.org/x/crypto/bcrypt"

// GeneratePassword 生成密码
func GeneratePassword(psw string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(psw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPassword 检查密码
// hashedPassword 已加密的密码
// password 待检查的密码
func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
