package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("super-secret-key") // Nên lưu trong biến môi trường thay vì hardcode

// GenerateJWT tạo JWT với email và thời gian sống 24 giờ
func GenerateJWT(email string) (string, error) {
	claims := &jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(), // Token hết hạn sau 24 giờ
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// VerifyJWT xác thực token và trả về email nếu hợp lệ
func VerifyJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra phương thức ký
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		email, ok := (*claims)["email"].(string)
		if !ok {
			return "", jwt.ErrInvalidKeyType
		}
		return email, nil
	}

	return "", jwt.ErrTokenInvalidId
}
