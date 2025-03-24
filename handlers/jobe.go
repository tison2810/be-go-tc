package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/models"
	"github.com/tison2810/be-go-tc/services"
)

func CheckJobeLanguages(c *fiber.Ctx) error {
	url := "http://jobe:80/jobe/index.php/restapi/languages"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error connecting to Jobe Server:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Cannot connect to Jobe Server")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error reading Jobe response")
	}

	return c.SendString(fmt.Sprintf("Supported Languages: %s", body))
}

func generateFileID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 8
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func UploadFileToJobeHandler(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Không thể đọc file từ request",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Không thể mở file",
		})
	}
	defer src.Close()

	fileContents, err := io.ReadAll(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.FileUploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Không thể đọc nội dung file: %v", err),
		})
	}

	fileID := generateFileID()

	base64Contents := base64.StdEncoding.EncodeToString(fileContents)

	requestData := models.UploadFileRequest{
		FileContents: base64Contents,
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.FileUploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Lỗi khi marshal JSON: %v", err),
		})
	}
	jobeServerURL := "http://jobe:80/jobe/index.php/restapi"
	url := fmt.Sprintf("%s/files/%s", jobeServerURL, fileID)
	log.Printf("Sending request to Jobe: %s", url) // Log URL

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.FileUploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Lỗi khi tạo request: %v", err),
		})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.FileUploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Lỗi khi gửi request tới Jobe: %v", err),
		})
	}
	defer resp.Body.Close()

	log.Printf("Jobe response status: %d", resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusNoContent: // 204
		return c.JSON(models.FileUploadResponse{
			Success: true,
			FileID:  fileID,
			Message: "File uploaded successfully to Jobe",
		})
	case http.StatusBadRequest: // 400
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Bad request - invalid file contents or missing parameters",
		})
	case http.StatusNotFound: // 404
		return c.Status(fiber.StatusNotFound).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Jobe server endpoint not found",
		})
	case http.StatusInternalServerError: // 500
		return c.Status(fiber.StatusInternalServerError).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Jobe server failed to write file to cache",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(models.FileUploadResponse{
			Success: false,
			Error:   fmt.Sprintf("Unexpected response code from Jobe: %d", resp.StatusCode),
		})
	}
}

func CheckFile(c *fiber.Ctx) error {
	fileID := c.Params("id")
	url := fmt.Sprintf("http://jobe:80/jobe/index.php/restapi/files/%s", fileID)
	resp, err := http.Head(url)
	if err != nil {
		log.Println("Error connecting to Jobe Server:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Cannot connect to Jobe Server")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent: // 204
		return c.SendString("File exists in Jobe cache")
	case http.StatusBadRequest: // 400
		return c.SendString("Missing fileID in url")
	case http.StatusNotFound: // 404
		return c.SendString("File not found in Jobe cache")
	default:
		return c.SendString(fmt.Sprintf("Unexpected response code from Jobe: %d", resp.StatusCode))
	}
}

func SubmitRun(c *fiber.Ctx) error {
	// Lấy jobeServerURL từ context
	jobeServerURL, ok := c.Locals("jobeServerURL").(string)
	if !ok {
		log.Println("Error: Cannot get jobeServerURL from context")
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  "Cannot get jobeServerURL from context",
		})
	}
	postIDStr := c.Query("post_id")
	studentMail := c.Locals("user_mail").(string)

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.SubmitRunResponse{
			Status: http.StatusBadRequest,
			Error:  "Invalid post_id",
		})
	}

	// Đọc dữ liệu từ request body
	var runSpec models.RunSpec
	if err := c.BodyParser(&runSpec); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.SubmitRunResponse{
			Status: http.StatusBadRequest,
			Error:  fmt.Sprintf("Invalid run_spec format: %v", err),
		})
	}

	// Kiểm tra các trường bắt buộc
	if runSpec.LanguageID == "" || runSpec.SourceCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.SubmitRunResponse{
			Status: http.StatusBadRequest,
			Error:  "Missing required fields: language_id and sourcecode are mandatory",
		})
	}

	// Tạo request data
	requestData := models.SubmitRunRequest{
		RunSpec: runSpec,
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  fmt.Sprintf("Error marshaling JSON: %v", err),
		})
	}

	// Gửi request POST tới Jobe server
	url := fmt.Sprintf("%s/runs", jobeServerURL)
	log.Printf("Sending request to Jobe: %s", url)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  fmt.Sprintf("Error creating request: %v", err),
		})
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  fmt.Sprintf("Error sending request to Jobe: %v", err),
		})
	}
	defer resp.Body.Close()

	log.Printf("Jobe response status: %d", resp.StatusCode)
	postService := services.NewPostService()
	// Xử lý response từ Jobe
	switch resp.StatusCode {
	case http.StatusOK: // 200: Run completed
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
				Status: http.StatusInternalServerError,
				Error:  fmt.Sprintf("Error reading Jobe response: %v", err),
			})
		}

		// Gọi CheckRunResult để kiểm tra và lưu kết quả
		studentRun, err := postService.CheckRunResult(postID, studentMail, string(body))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
				Status: http.StatusInternalServerError,
				Error:  fmt.Sprintf("Error checking run result: %v", err),
			})
		}

		return c.JSON(models.SubmitRunResponse{
			Status: http.StatusOK,
			Result: string(body),
			Score:  studentRun.Score,
			Log:    studentRun.Log,
		})
	case http.StatusAccepted: // 202: Job queued for later execution
		return c.Status(http.StatusAccepted).JSON(models.SubmitRunResponse{
			Status: http.StatusAccepted,
			Result: "Job queued for later execution",
		})
	case http.StatusBadRequest: // 400: Bad request
		return c.Status(fiber.StatusBadRequest).JSON(models.SubmitRunResponse{
			Status: http.StatusBadRequest,
			Error:  "Bad request - invalid run_spec or missing parameters",
		})
	case http.StatusNotFound: // 404: Not found
		return c.Status(fiber.StatusNotFound).JSON(models.SubmitRunResponse{
			Status: http.StatusNotFound,
			Error:  "Jobe server endpoint not found",
		})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  fmt.Sprintf("Unexpected response code from Jobe: %d", resp.StatusCode),
		})
	}
}
