package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/models"
	"github.com/tison2810/be-go-tc/utils"
)

// PostService chứa các phương thức liên quan đến post
var flaskClient *utils.FlaskClient

func init() {
	flaskClient = utils.NewFlaskClient()
}

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

func GetPostStats(userMail string, postIDs []uuid.UUID) []models.PostStats {
	var stats []models.PostStats
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

func CreatePostFormData(c *fiber.Ctx) (*models.Post, error) {
	post := new(models.Post)

	// Lấy email từ context
	post.UserMail, _ = c.Locals("email").(string)
	if post.UserMail == "" {
		return nil, fmt.Errorf("user email not found in context")
	}

	post.Title = c.FormValue("title")
	post.Description = c.FormValue("description")
	post.Subject = "KTLT"

	testcase := new(models.Testcase)
	file, err := c.FormFile("input")
	if err == nil { // File tồn tại
		// Mở file
		fileHandle, err := file.Open()
		if err != nil {
			log.Printf("Failed to open uploaded file: %v", err)
			return nil, fmt.Errorf("failed to open uploaded file: %v", err)
		}
		defer fileHandle.Close()

		// Đọc nội dung file
		fileContent, err := io.ReadAll(fileHandle)
		if err != nil {
			log.Printf("Failed to read uploaded file: %v", err)
			return nil, fmt.Errorf("failed to read uploaded file: %v", err)
		}

		testcase.Input = string(fileContent)
	} else if err != fiber.ErrBadRequest { // Chỉ log lỗi nếu không phải trường hợp file không được gửi
		log.Printf("Failed to get supfile from form: %v", err)
	}
	testcase.Expected = c.FormValue("expected")
	testCode := c.FormValue("code")
	headers := `#include "main.h"
	#include "tc.h"
	#include "hcmcampaign.h"
	
	`
	testcase.Code = headers + testCode
	if testcase.Input != "" || testcase.Expected != "" || testcase.Code != "" {
		post.Testcase = testcase
	}

	post.ID = uuid.New()
	post.CreatedAt = time.Now()
	post.LastModified = post.CreatedAt

	if post.Testcase != nil {
		post.Testcase.PostID = post.ID
	}

	if err := database.DB.Db.Create(&post).Error; err != nil {
		return nil, fmt.Errorf("failed to save post and testcase: %v", err)
	}

	// Upload testcase input lên Jobe server nếu có
	if post.Testcase != nil && post.Testcase.Input != "" {
		go func() {
			fileName := strings.ReplaceAll(post.ID.String(), "-", "")
			fileContents := []byte(post.Testcase.Input)
			base64Contents := base64.StdEncoding.EncodeToString(fileContents)

			requestData := models.UploadFileRequest{
				FileContents: base64Contents,
			}
			jsonData, err := json.Marshal(requestData)
			if err != nil {
				log.Printf("Failed to marshal JSON for Jobe: %v", err)
				return
			}

			url := fmt.Sprintf("http://jobe:80/jobe/index.php/restapi/files/%s", fileName)
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Failed to create Jobe request: %v", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to send request to Jobe: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusNoContent {
				log.Printf("Jobe returned unexpected status: %d", resp.StatusCode)
			}
		}()
	}

	// Gọi Flask để trace nếu cần
	go func() {
		trace, err := flaskClient.CallTrace(post.ID.String())
		if err != nil {
			log.Printf("Failed to call Flask trace API: %v", err)
			return
		}
		post.Trace = trace
		if err := database.DB.Db.Save(&post).Error; err != nil {
			log.Printf("Failed to update post trace: %v", err)
		}
	}()

	return post, nil
}

type HotPost struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Title    string    `json:"title" gorm:"type:varchar(255);not null"`
	Author   string    `json:"author" gorm:"type:varchar(100);not null"`
	HotScore int       `json:"hot_score" gorm:"type:int;default:0"`
}

func GetHotPost() []HotPost {
	var hotPosts []HotPost
	database.DB.Db.Raw(`
		SELECT p.id, (COALESCE(i.interaction_count, 0) + COALESCE(c.comment_count, 0)) as hot_score, p.title, COALESCE(u.last_name || ' ' || u.first_name, '') AS author
		FROM posts p
		LEFT JOIN (
			SELECT post_id, COUNT(*) as interaction_count
			FROM interactions
			GROUP BY post_id
		) i ON p.id = i.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) as comment_count
			FROM comments
			GROUP BY post_id
		) c ON p.id = c.post_id
		LEFT JOIN users u ON p.user_mail = u.mail
		WHERE p.subject = 'KTLT' AND p.post_status IN ('active', 'similar')
		ORDER BY hot_score DESC, p.created_at DESC
		LIMIT 3
		`).Scan(&hotPosts)
	if len(hotPosts) == 0 {
		return []HotPost{}
	}
	return hotPosts
}
