package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FlaskClient là struct để quản lý kết nối tới Flask API
type FlaskClient struct {
	agent   *fiber.Agent
	baseURL string
}

// NewFlaskClient tạo một instance mới của FlaskClient
func NewFlaskClient() *FlaskClient {
	baseURL := "http://flask:5000" // Tên service trong docker-compose.yml
	agent := fiber.AcquireAgent()
	agent.Timeout(time.Second * 5)

	return &FlaskClient{
		agent:   agent,
		baseURL: baseURL,
	}
}

// CallTrace gọi endpoint /trace của Flask API và trả về giá trị trace
func (fc *FlaskClient) CallTrace(postID string) (string, error) {
	req := fc.agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(fc.baseURL + "/trace")
	req.Header.SetContentType("application/json")

	// Tạo request body
	requestBody, err := json.Marshal(fiber.Map{
		"post_id": postID,
	})
	if err != nil {
		log.Printf("Failed to marshal request body: %v", err)
		return "", err
	}
	req.SetBody(requestBody)

	// Gửi request
	if err := fc.agent.Parse(); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return "", err
	}

	// Lấy response body
	status, body, errs := fc.agent.Bytes()
	if len(errs) > 0 {
		log.Printf("Failed to call Flask API: %v", errs)
		return "", errs[0]
	}

	// Kiểm tra status code
	if status != fiber.StatusOK {
		log.Printf("Flask API returned non-200 status: %d, body: %s", status, string(body))
		return "", fmt.Errorf("Flask API returned status: %d", status)
	}

	// Parse JSON response
	var response struct {
		PostID string `json:"post_id"`
		Trace  string `json:"trace"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Failed to parse Flask API response: %v", err)
		return "", err
	}

	// Trả về giá trị trace
	return response.Trace, nil
}
