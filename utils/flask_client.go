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

func (fc *FlaskClient) CallSuggest(email string) ([]string, error) {
	req := fc.agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(fc.baseURL + "/suggest")
	req.Header.SetContentType("application/json")

	// Tạo request body
	requestBody, err := json.Marshal(fiber.Map{
		"student_email": email,
		"subject":       "KTLT",
	})
	if err != nil {
		log.Printf("Failed to marshal request body: %v", err)
		return nil, err
	}
	req.SetBody(requestBody)

	// Gửi request
	if err := fc.agent.Parse(); err != nil {
		log.Printf("Failed to parse request: %v", err)
		return nil, err
	}

	// Lấy response body
	status, body, errs := fc.agent.Bytes()
	if len(errs) > 0 {
		log.Printf("Failed to call Flask API: %v", errs)
		return nil, errs[0]
	}

	// Kiểm tra status code
	if status != fiber.StatusOK {
		log.Printf("Flask API returned non-200 status: %d, body: %s", status, string(body))
		return nil, fmt.Errorf("flask API returned status: %d", status)
	}

	// Parse JSON response
	var response struct {
		InterestLabel  string   `json:"interest_label"`
		SuggestedPosts []string `json:"suggested_posts"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Failed to parse Flask API response: %v", err)
		return nil, err
	}

	return response.SuggestedPosts, nil
}

type SimilarPost struct {
	PostID      string  `json:"post_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Author      string  `json:"author"`
	Input       string  `json:"input"`
	Expected    string  `json:"expected"`
	Similarity  float64 `json:"similarity"`
}

func (fc *FlaskClient) CallSimilarPost(postID string) ([]SimilarPost, error) {
	req := fc.agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(fc.baseURL + "/similar_post")
	req.Header.SetContentType("application/json")

	requestBody, err := json.Marshal(fiber.Map{
		"post_id": postID,
	})
	if err != nil {
		return nil, err
	}
	req.SetBody(requestBody)

	if err := fc.agent.Parse(); err != nil {
		return nil, err
	}

	status, body, errs := fc.agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if status == fiber.StatusNotFound {
		return nil, nil // Không tìm thấy bài viết tương tự
	}
	if status != fiber.StatusOK {
		return nil, fmt.Errorf("flask API returned status: %d", status)
	}

	var response struct {
		SimilarPosts []SimilarPost `json:"similar_posts"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	log.Printf("Similar posts: %v", response.SimilarPosts)
	return response.SimilarPosts, nil
}

func (fc *FlaskClient) CallRelatedPost(postID string) ([]SimilarPost, error) {
	req := fc.agent.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(fc.baseURL + "/related_post")
	req.Header.SetContentType("application/json")

	requestBody, err := json.Marshal(fiber.Map{
		"post_id": postID,
	})
	if err != nil {
		return nil, err
	}
	req.SetBody(requestBody)

	if err := fc.agent.Parse(); err != nil {
		return nil, err
	}

	status, body, errs := fc.agent.Bytes()
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if status == fiber.StatusNotFound {
		return nil, nil // Không tìm thấy bài viết tương tự
	}
	if status != fiber.StatusOK {
		return nil, fmt.Errorf("flask API returned status: %d", status)
	}

	var response struct {
		RelatedPosts []SimilarPost `json:"related_posts"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	log.Printf("Related posts: %v", response.RelatedPosts)
	return response.RelatedPosts, nil
}
