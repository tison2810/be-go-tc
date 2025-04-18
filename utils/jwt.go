package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// var jwtKey = []byte(os.Getenv("JWT_SECRET"))
var jwtKey = []byte("super-secret-key")

// GenerateJWT tạo JWT với email và thời gian sống 24 giờ
func GenerateJWT(email string, role string) (string, error) {
	claims := &jwt.MapClaims{
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"iss":   "login",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// VerifyJWT xác thực token và trả về email nếu hợp lệ
func VerifyJWT(tokenString string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtKey, nil
	})

	if err != nil {
		// Phân loại lỗi cụ thể hơn
		if err == jwt.ErrSignatureInvalid {
			return "", "", fmt.Errorf("invalid token signature")
		}
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", "", fmt.Errorf("token has expired")
		}
		return "", "", fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		if email, ok := (*claims)["email"].(string); ok {
			if role, ok := (*claims)["role"].(string); ok {
				return email, role, nil
			}
		}
		return "", "", fmt.Errorf("invalid claims")
	}
	return "", "", fmt.Errorf("invalid")
}
