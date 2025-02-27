package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/models"
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

func UploadFileToJobe(c *fiber.Ctx) error {
	// Nhận file từ request
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("File upload failed")
	}

	// Mở file để đọc nội dung
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Cannot open uploaded file")
	}
	defer src.Close()

	// Mã hóa nội dung file thành Base64
	encodedData, err := encodeFileToBase64(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Encoding to Base64 failed")
	}

	// Gửi file lên Jobe và nhận lại file_id
	fileID, err := postBase64FileToJobe(file.Filename, encodedData)
	if err != nil {
		log.Println("Upload to Jobe failed:", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Upload to Jobe failed")
	}

	// Trả về file_id cho client
	return c.JSON(fiber.Map{"file_id": fileID})
}

// Mã hóa file thành Base64
func encodeFileToBase64(src io.Reader) (string, error) {
	// Đọc toàn bộ nội dung file
	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// Mã hóa thành Base64
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

// Gửi nội dung file đã mã hóa Base64 lên Jobe và nhận lại file_id
func postBase64FileToJobe(filename, base64Content string) (string, error) {
	jobeURL := "http://jobe:80/jobe/index.php/restapi/files"

	// Chuẩn bị payload JSON
	payload := map[string]string{
		"file_contents": base64Content,
		"filename":      filename,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Gửi request POST đến Jobe
	resp, err := http.Post(jobeURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Đọc phản hồi từ Jobe
	body, _ := io.ReadAll(resp.Body)
	log.Println("Jobe Response:", string(body)) // Debug Response

	// Parse JSON response để lấy file_id
	var result models.UploadResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error parsing Jobe response: %s", body)
	}

	return result.FileID, nil
}
