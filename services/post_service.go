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

func GetTestcaseByPostID(postID uuid.UUID) (models.Testcase, error) {
	var testcase models.Testcase
	if err := database.DB.Db.Where("post_id = ?", postID).First(&testcase).Error; err != nil {
		return models.Testcase{}, err
	}
	return testcase, nil
}

type PostStats struct {
	PostID       uuid.UUID  `gorm:"column:post_id"`
	LikeCount    int64      `gorm:"column:like_count"`
	CommentCount int64      `gorm:"column:comment_count"`
	LikeID       *uuid.UUID `gorm:"column:like_id"` // ID của like nếu user đã like
}

func GetPostStats(userMail string, postIDs []uuid.UUID) []PostStats {
	var stats []PostStats
	if len(postIDs) == 0 {
		return stats
	}

	database.DB.Db.Raw(`
        SELECT 
            i.post_id,
            SUM(CASE WHEN i.type = 'Like' AND i.is_like = true THEN 1 ELSE 0 END) as like_count,
            SUM(CASE WHEN i.type = 'Comment' THEN 1 ELSE 0 END) as comment_count,
            MAX(CASE WHEN i.user_mail = ? AND i.type = 'Like' AND i.is_like = true THEN i.id END) as like_id
        FROM interactions i
        WHERE i.post_id IN (?)
        GROUP BY i.post_id
    `, userMail, postIDs).Scan(&stats)
	return stats
}
