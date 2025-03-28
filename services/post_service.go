package services

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/database" // Import package database
	"github.com/tison2810/be-go-tc/models"
)

// PostService chứa các phương thức liên quan đến post
type PostService struct {
	// Không cần field DB nữa vì sẽ dùng database.DB trực tiếp
}

// NewPostService tạo một PostService mới
func NewPostService() *PostService {
	return &PostService{}
}

// CheckRunResult kiểm tra kết quả chạy code từ Jobe và so sánh với expected của testcase
func (s *PostService) CheckRunResult(
	postID uuid.UUID,
	studentMail string,
	jobeResponse string,
) (*models.StudentRunTestcase, error) {
	// 1. Parse response từ Jobe
	var jobeResult models.JobeRunResult
	if err := json.Unmarshal([]byte(jobeResponse), &jobeResult); err != nil {
		return nil, err
	}

	// 2. Lấy testcase từ database
	var testcase models.Testcase
	if err := database.DB.Db.Where("post_id = ?", postID).First(&testcase).Error; err != nil {
		return nil, err
	}

	// 3. So sánh stdout với expected
	// Chuẩn hóa stdout và expected (loại bỏ khoảng trắng thừa, xuống dòng)
	stdout := strings.TrimSpace(jobeResult.Stdout)
	expected := strings.TrimSpace(testcase.Expected)

	// Tính score: 1 nếu khớp, 0 nếu không khớp
	score := 0
	logMessage := stdout
	if stdout == expected {
		score = 1
	}

	// Kiểm tra lỗi biên dịch hoặc runtime
	if jobeResult.Cmpinfo != "" {
		logMessage = "Compilation error: " + jobeResult.Cmpinfo
		score = 0
	} else if jobeResult.Stderr != "" {
		logMessage = "Runtime error: " + jobeResult.Stderr
		score = 0
	} else if jobeResult.Outcome != 15 {
		logMessage = "Execution failed: " + strconv.Itoa(jobeResult.Outcome)
		score = 0
	}
	studentRun := models.StudentRunTestcase{
		ID:          uuid.New(),
		PostID:      postID,
		StudentMail: studentMail,
		Log:         logMessage,
		Score:       score,
	}

	if err := database.DB.Db.Create(&studentRun).Error; err != nil {
		return nil, err
	}

	return &studentRun, nil
}
