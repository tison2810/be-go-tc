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
	PostID              uuid.UUID  `gorm:"column:post_id"`
	LikeCount           int64      `gorm:"column:like_count"`
	CommentCount        int64      `gorm:"column:comment_count"`
	LikeID              *uuid.UUID `gorm:"column:like_id"`
	VerifiedTeacherMail *string    `gorm:"column:verified_teacher_mail"`
	Views               int        `gorm:"column:views"`
	Runs                int        `gorm:"column:runs"`
}

func GetPostStats(userMail string, postIDs []uuid.UUID) []PostStats {
	var stats []PostStats
	if len(postIDs) == 0 {
		return stats
	}

	database.DB.Db.Raw(`
        SELECT 
            p.id as post_id,
            p.views,
            p.runs,
            COALESCE((
                SELECT COUNT(*) 
                FROM interactions i 
                WHERE i.post_id = p.id AND i.is_like = true
            ), 0) as like_count,
            COALESCE((
                SELECT COUNT(*) 
                FROM comments c 
                WHERE c.post_id = p.id AND c.is_deleted = false
            ), 0) as comment_count,
            MAX(CASE WHEN i.user_mail = ? AND i.is_like = true THEN i.id::text END)::uuid as like_id,
            tvp.teacher_mail as verified_teacher_mail
        FROM posts p
        LEFT JOIN interactions i ON p.id = i.post_id
        LEFT JOIN teacher_verify_posts tvp ON p.id = tvp.post_id
        WHERE p.id IN (?)
        GROUP BY p.id, p.views, p.runs, tvp.teacher_mail
    `, userMail, postIDs).Scan(&stats)
	return stats
}
