package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/models"
	"github.com/tison2810/be-go-tc/services"
	"github.com/tison2810/be-go-tc/utils"
)

func generateFileID(tokenString string) string {
	email, err := utils.VerifyJWT(tokenString)
	if err != nil {
		return ""
	}
	dbUser, err := services.FindUserByEmail(database.DB.Db, email)
	if err != nil {
		panic(fmt.Errorf("failed to find user by email: %w", err))
	}
	return dbUser.Maso
}

func UploadTwoFilesHandler(c *fiber.Ctx) error {
	// Lấy file .cpp
	cppFile, err := c.FormFile("cpp_file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Không thể đọc file .cpp từ request",
		})
	}
	if !strings.HasSuffix(cppFile.Filename, ".cpp") {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "File cpp_file phải có đuôi .cpp",
		})
	}

	// Lấy file .h
	hFile, err := c.FormFile("h_file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "Không thể đọc file .h từ request",
		})
	}
	if !strings.HasSuffix(hFile.Filename, ".h") {
		return c.Status(fiber.StatusBadRequest).JSON(models.FileUploadResponse{
			Success: false,
			Error:   "File h_file phải có đuôi .h",
		})
	}

	baseID := generateFileID(c.Locals("token").(string))
	// baseID := "2112198"
	cppFileID := fmt.Sprintf("%scpp", baseID) // [xxxxxxx]cpp
	hFileID := fmt.Sprintf("%sh", baseID)     // [xxxxxxx]h

	c.Locals("form_file_key", "cpp_file")
	c.Locals("file_id", cppFileID)
	if err := UploadFileToJobeHandler(c); err != nil {
		return err
	}

	// Reset response
	c.Response().Reset()

	c.Locals("form_file_key", "h_file")
	c.Locals("file_id", hFileID)
	if err := UploadFileToJobeHandler(c); err != nil {
		return err
	}

	// Trả về kết quả thành công cho cả hai file
	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Cả hai file đã được upload thành công",
		"cpp_file_id": cppFileID,
		"h_file_id":   hFileID,
	})
}

func RunCode(c *fiber.Ctx) error {

	jobeServerURL := "http://jobe:80/jobe/index.php/restapi"

	// Lấy email từ Locals (do AuthMiddleware cung cấp)
	studentMail, ok := c.Locals("email").(string)
	if !ok || studentMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(models.SubmitRunResponse{
			Status: http.StatusUnauthorized,
			Error:  "User email not found in context",
		})
	}

	// Lấy studentID từ database
	studentID, err := services.GetMaso(database.DB.Db, studentMail)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  fmt.Sprintf("Error getting student ID: %v", err),
		})
	}

	// Lấy post_id từ param
	postIDStr := c.Params("id")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.SubmitRunResponse{
			Status: http.StatusBadRequest,
			Error:  "Missing post_id in URL parameter",
		})
	}
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.SubmitRunResponse{
			Status: http.StatusBadRequest,
			Error:  "Invalid post_id",
		})
	}

	// Lấy testcase từ database
	testcase, err := services.GetTestcaseByPostID(postID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.SubmitRunResponse{
			Status: http.StatusNotFound,
			Error:  fmt.Sprintf("Error retrieving testcase: %v", err),
		})
	}

	mainCode := testcase.Code

	var isSuggested bool
	suggestedPosts, err := flaskClient.CallSuggest(studentMail)
	if err != nil {
		log.Printf("Failed to call Flask suggest API: %v", err)
	} else {
		for _, sugID := range suggestedPosts {
			if sugID == postID.String() {
				isSuggested = true
				break
			}
		}
	}

	// Cập nhật RunPosts và RunSuggestedPosts trong user
	var user models.User
	if err := database.DB.Db.First(&user, "mail = ?", studentMail).Error; err != nil {
		log.Printf("User not found: %v", err)
	} else {
		user.RunPosts++ // Luôn tăng RunPosts
		if isSuggested {
			user.RunSuggestedPosts++ // Tăng RunSuggestedPosts nếu là suggested
		}
		if err := database.DB.Db.Save(&user).Error; err != nil {
			log.Printf("Failed to update user RunPosts: %v", err)
		}
	}

	// Tạo config filename (cố định hoặc dùng post_id tùy ý)
	configFileName := "config.txt" // Giữ cố định như trong code của bạn
	fileName := strings.ReplaceAll(postID.String(), "-", "")
	// Tạo RunSpec với cấu hình cứng
	// 	runSpec := models.RunSpec{
	// 		LanguageID: "cpp",
	// 		SourceCode: fmt.Sprintf(`#include "hcmcampaign.h"

	// using namespace std;

	// void g_satc_01() {
	//     cout << "----- Sample Testcase 01 -----" << endl;
	//     Configuration* config = new Configuration("%s");
	//     cout << config->str() << endl;
	//     delete config;
	// }

	// int main(int argc, const char * argv[]) {
	//     g_satc_01();
	//     return 0;
	// }`, configFileName),
	// 		SourceFilename: "main.cpp",
	// 		Input:          "",
	// 		FileList: [][]interface{}{
	// 			{fmt.Sprintf("%scpp", studentID), "hcmcampaign.cpp"}, // Dùng studentID cho file_id
	// 			{fmt.Sprintf("%sh", studentID), "hcmcampaign.h"},     // Dùng studentID cho file_id
	// 			{"systemmainh", "main.h"},
	// 			{"configtxt", configFileName}, // File config cố định
	// 		},
	// 		Parameters: map[string]interface{}{
	// 			"max_execution_time": 5,
	// 			"max_memory_usage":   1000000,
	// 			"compileargs":        []string{"-I .", "-std=c++11"},
	// 			"linkargs":           []string{"hcmcampaign.cpp"},
	// 			"args":               []string{configFileName},
	// 		},
	// 		Debug: true,
	// 	}

	runSpec := models.RunSpec{
		LanguageID:     "cpp",
		SourceCode:     mainCode,
		SourceFilename: "main.cpp",
		Input:          "",
		FileList: [][]interface{}{
			{fmt.Sprintf("%scpp", studentID), "hcmcampaign.cpp"}, // Dùng studentID cho file_id
			{fmt.Sprintf("%sh", studentID), "hcmcampaign.h"},     // Dùng studentID cho file_id
			{"systemmainh", "main.h"},
			{fileName, configFileName}, // File config cố định
		},
		Parameters: map[string]interface{}{
			"max_execution_time": 5,
			"max_memory_usage":   1000000,
			"compileargs":        []string{"-I .", "-std=c++11"},
			"linkargs":           []string{"hcmcampaign.cpp"},
			"args":               []string{configFileName},
		},
		Debug: true,
	}

	requestData := models.SubmitRunRequest{
		RunSpec: runSpec,
	}
	log.Printf("RunSpec: %+v", requestData)

	// Chuyển thành JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  fmt.Sprintf("Error marshaling JSON: %v", err),
		})
	}

	log.Printf("Sending request to Jobe with data: %s", string(jsonData))

	url := fmt.Sprintf("%s/runs", jobeServerURL)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Jobe response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
			Status: http.StatusInternalServerError,
			Error:  fmt.Sprintf("Error reading Jobe response: %v", err),
		})
	}

	log.Printf("Jobe response status: %d", resp.StatusCode)
	log.Printf("Jobe response body: %s", string(body))

	// Khởi tạo PostService
	postService := services.NewPostService()

	// Xử lý response từ Jobe
	switch resp.StatusCode {
	case http.StatusOK:
		var jobeResult models.JobeRunResult
		if err := json.Unmarshal(body, &jobeResult); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.SubmitRunResponse{
				Status: http.StatusInternalServerError,
				Error:  fmt.Sprintf("Error parsing Jobe response: %v, body: %s", err, string(body)),
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

		// Trả về chỉ stdout trong Result
		return c.JSON(models.SubmitRunResponse{
			Status: http.StatusOK,
			Result: jobeResult.Stdout,
			Score:  studentRun.Score,
			Log:    studentRun.Log,
		})
	case http.StatusAccepted: // 202: Job queued
		return c.Status(fiber.StatusAccepted).JSON(models.SubmitRunResponse{
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
