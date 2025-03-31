package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("super-secret-key") // Nên lưu trong biến môi trường thay vì hardcode

// GenerateJWT tạo JWT với email và thời gian sống 24 giờ
func GenerateJWT(email string) (string, error) {
	claims := &jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(), // Thêm thời điểm phát hành
		"iss":   "my-app",          // Thêm issuer (tùy chỉnh)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// VerifyJWT xác thực token và trả về email nếu hợp lệ
func VerifyJWT(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})

	if err != nil {
		// Phân loại lỗi cụ thể hơn
		if err == jwt.ErrSignatureInvalid {
			return "", fmt.Errorf("invalid token signature")
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", fmt.Errorf("token has expired")
		}
		return "", fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		if email, ok := (*claims)["email"].(string); ok {
			return email, nil
		}
		return "", fmt.Errorf("email claim is missing or invalid")
	}

	return "", fmt.Errorf("invalid token claims")
}
