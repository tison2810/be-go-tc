package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/models"
	"github.com/tison2810/be-go-tc/services"
	"github.com/tison2810/be-go-tc/utils"
	"gorm.io/gorm"
)

var flaskClient *utils.FlaskClient

func init() {
	flaskClient = utils.NewFlaskClient()
}

// @Summary Create a new post
// @Tags Post
// @Accept json
// @Produce json
// @Param post body models.Post true "Post content"
// @Success 201 {object} models.Post
// @Security BearerAuth
// @Router /create [post]

func CreatePost(c *fiber.Ctx) error {
	post := new(models.Post)

	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}
	post.UserMail, _ = c.Locals("email").(string)
	if post.UserMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}
	post.ID = uuid.New()
	post.LastModified = time.Now()
	post.Subject = "KTLT"

	if post.Testcase != nil {
		post.Testcase.PostID = post.ID
	}

	if err := database.DB.Db.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save post and testcase: " + err.Error(),
		})
	}

	if post.Testcase != nil && post.Testcase.Input != "" {
		go func() {
			// Tạo tên file từ UUID không dấu gạch nối
			fileName := strings.ReplaceAll(post.ID.String(), "-", "") // Ví dụ: 6ce9aea776a141d1a92b7faa12ecae20.txt
			fileContents := []byte(post.Testcase.Input)               // Nội dung file từ Testcase.Input

			// Mã hóa base64
			base64Contents := base64.StdEncoding.EncodeToString(fileContents)

			// Tạo request data
			requestData := models.UploadFileRequest{
				FileContents: base64Contents,
			}
			jsonData, err := json.Marshal(requestData)
			if err != nil {
				log.Printf("Failed to marshal JSON for Jobe: %v", err)
				return
			}

			// Gửi request tới Jobe server
			jobeServerURL := "http://jobe:80/jobe/index.php/restapi"
			url := fmt.Sprintf("%s/files/%s", jobeServerURL, fileName)
			log.Printf("Sending request to Jobe: %s", url)

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

	go func() {
		trace, err := flaskClient.CallTrace(post.ID.String())
		if err != nil {
			log.Printf("Failed to call Flask trace API: %v", err)
			return
		}
		post.Trace = trace
		// fmt.Print(post.Trace)
	}()
	return c.Status(fiber.StatusCreated).JSON(post)
}

func CreatePostFormData(c *fiber.Ctx) error {
	post := new(models.Post)

	// Lấy email từ context (giữ nguyên)
	post.UserMail, _ = c.Locals("email").(string)
	if post.UserMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Lấy dữ liệu từ form-data
	post.Title = c.FormValue("title")
	post.Description = c.FormValue("description")
	post.Subject = "KTLT" // Hardcode như hiện tại

	// Tạo Testcase từ form-data
	testcase := new(models.Testcase)
	testcase.Input = c.FormValue("input")
	testcase.Expected = c.FormValue("expected")
	testcase.Code = c.FormValue("code")
	if testcase.Input != "" || testcase.Expected != "" || testcase.Code != "" {
		post.Testcase = testcase
	}

	// Tạo UUID cho post
	post.ID = uuid.New()
	post.LastModified = time.Now()

	if post.Testcase != nil {
		post.Testcase.PostID = post.ID
	}

	// Lưu post vào database
	if err := database.DB.Db.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save post and testcase: " + err.Error(),
		})
	}

	// Gửi Testcase.Input tới Jobe (nếu có)
	if post.Testcase != nil && post.Testcase.Input != "" {
		go func() {
			// Tạo tên file từ UUID không dấu gạch nối
			fileName := strings.ReplaceAll(post.ID.String(), "-", "") + ".txt" // Ví dụ: 6ce9aea776a141d1a92b7faa12ecae20.txt
			fileContents := []byte(post.Testcase.Input)                        // Nội dung file từ Testcase.Input (giữ xuống dòng thực tế)

			// Mã hóa base64
			base64Contents := base64.StdEncoding.EncodeToString(fileContents)

			// Tạo request data
			requestData := models.UploadFileRequest{
				FileContents: base64Contents,
			}
			jsonData, err := json.Marshal(requestData)
			if err != nil {
				log.Printf("Failed to marshal JSON for Jobe: %v", err)
				return
			}

			// Gửi request tới Jobe server
			jobeServerURL := "http://jobe:80/jobe/index.php/restapi"
			url := fmt.Sprintf("%s/files/%s", jobeServerURL, fileName)
			log.Printf("Sending request to Jobe: %s", url)

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

	// Gọi Flask API để lấy trace
	go func() {
		trace, err := flaskClient.CallTrace(post.ID.String())
		if err != nil {
			log.Printf("Failed to call Flask trace API: %v", err)
			return
		}
		post.Trace = trace
		if err := database.DB.Db.Save(&post).Error; err != nil {
			log.Printf("Failed to update post trace: %v", err)
			return
		}
		// fmt.Print(post.Trace)
	}()

	// Trả về JSON
	return c.Status(fiber.StatusCreated).JSON(post)
}

func GetAllPostsID(c *fiber.Ctx) error {
	var posts []models.Post
	database.DB.Db.Find(&posts)
	var postsID []uuid.UUID
	for _, post := range posts {
		postsID = append(postsID, post.ID)
	}
	return c.Status(fiber.StatusOK).JSON(postsID)
}

// func GetAllPosts(c *fiber.Ctx) error {
// 	var posts []models.Post
// 	database.DB.Db.Preload("Testcase").Find(&posts)
// 	return c.Status(fiber.StatusOK).JSON(posts)
// }

func GetAllPosts(c *fiber.Ctx) error {
	// Lấy email từ Locals (để dùng trong GetPostStats)
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Lấy tất cả bài đăng
	var posts []models.Post
	if err := database.DB.Db.Preload("Testcase").Where("is_deleted = ?", false).Find(&posts).Error; err != nil {
		log.Printf("Failed to fetch posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch posts",
		})
	}

	// Lấy danh sách post IDs
	var postIDs []uuid.UUID
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	// Lấy thông tin tương tác từ GetPostStats
	stats := services.GetPostStats(email, postIDs)
	statsMap := make(map[uuid.UUID]services.PostStats)
	for _, stat := range stats {
		statsMap[stat.PostID] = stat
	}

	// Lấy thông tin user để tạo Author
	var users []models.User
	if err := database.DB.Db.Find(&users).Error; err != nil {
		log.Printf("Failed to fetch users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[user.Mail] = user.LastName + " " + user.FirstName
	}

	// Tạo danh sách kết quả dạng PostWithType
	var resultPosts []PostWithType
	for _, post := range posts {
		stat, exists := statsMap[post.ID]
		if !exists {
			stat = services.PostStats{}
		}

		// Tạo Author từ userMap
		author := userMap[post.UserMail]
		if author == "" {
			author = "Unknown"
		}

		// Mặc định PostType là 0 (ngẫu nhiên) vì không có ngữ cảnh cụ thể
		postType := 0

		resultPosts = append(resultPosts, PostWithType{
			Post:     post,
			Author:   author,
			PostType: postType,
			Interaction: InteractionInfo{
				LikeCount:    stat.LikeCount,
				CommentCount: stat.CommentCount,
				Views:        stat.Views,
				Runs:         stat.Runs,
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(resultPosts)
}

func GetPost(c *fiber.Ctx) error {
	id := c.Params("id")
	post := new(models.Post)
	testcase := new(models.Testcase)
	database.DB.Db.Where("id = ?", id).First(&post)
	if post.ID == uuid.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}
	database.DB.Db.Where("post_id = ?", id).First(&testcase)
	post.Testcase = testcase
	return c.Status(fiber.StatusOK).JSON(post)
}

func GetTop5Post(c *fiber.Ctx) error {
	var posts []models.Post
	database.DB.Db.Limit(5).Order("created_at desc").Find(&posts)
	return c.Status(fiber.StatusOK).JSON(posts)
}

func UpdatePost(c *fiber.Ctx) error {
	id := c.Params("id")
	post := new(models.Post)
	result := database.DB.Db.Where("id = ?", id).Preload("Testcase").First(&post)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	post.LastModified = time.Now()
	if err := database.DB.Db.Save(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update post",
		})
	}

	if post.Testcase != nil {
		post.Testcase.PostID = post.ID
		if err := database.DB.Db.Save(&post.Testcase).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update testcase",
			})
		}
	} else {
		newTestcase := models.Testcase{
			PostID:   post.ID,
			Input:    post.Testcase.Input,
			Expected: post.Testcase.Expected,
			Code:     post.Testcase.Code,
		}
		if err := database.DB.Db.Create(&newTestcase).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create new testcase",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(post)
}

func DeletePost(c *fiber.Ctx) error {
	id := c.Params("id")
	post := new(models.Post)
	database.DB.Db.Where("id = ?", id).First(&post)
	if post.ID == uuid.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	database.DB.Db.Update("is_deleted", true).Where("id = ?", id).First(&post)
	return c.Status(fiber.StatusNoContent).Send(nil)
}

func LikePost(c *fiber.Ctx) error {
	// Lấy email từ Locals (do AuthMiddleware cung cấp)
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Lấy post_id từ params
	postIDStr := c.Params("id")
	if postIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing post_id in URL parameter",
		})
	}
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid post_id",
		})
	}

	// Dùng transaction để xử lý toggle like
	var interaction models.Interaction
	err = database.DB.Db.Transaction(func(tx *gorm.DB) error {
		// Kiểm tra xem interaction đã tồn tại chưa
		if err := tx.Where("post_id = ? AND user_mail = ?", postID, userMail).First(&interaction).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Nếu chưa tồn tại, tạo mới với IsLike = true
				interaction = models.Interaction{
					ID:        uuid.New(),
					PostID:    postID,
					UserMail:  userMail,
					CreatedAt: time.Now(),
					IsLike:    true,
				}
				return tx.Create(&interaction).Error
			}
			return err
		}

		// Nếu đã tồn tại, toggle IsLike
		interaction.IsLike = !interaction.IsLike
		return tx.Save(&interaction).Error
	})

	if err != nil {
		log.Printf("Failed to process like: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process like: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(interaction)
}

func GetLikeCount(postID uuid.UUID) int64 {
	var likeCount int64
	database.DB.Db.Model(&models.Interaction{}).
		Where("post_id = ? AND type = ? AND is_like = ?", postID, "Like", true).
		Count(&likeCount)
	return likeCount
}

type InteractionInfo struct {
	LikeCount           int64      `json:"like_count"`    // Số lượt like
	CommentCount        int64      `json:"comment_count"` // Số lượt comment
	LikeID              *uuid.UUID `json:"like_id"`       // ID của like nếu user đã like, null nếu chưa
	VerifiedTeacherMail *string    `json:"verified_teacher_mail"`
	Views               int        `json:"view_count"` // Số lượt xem
	Runs                int        `json:"run_count"`  // Số lượt chạy
}
type PostWithType struct {
	models.Post
	Author      string          `json:"author"`      // Tên tác giả
	PostType    int             `json:"post_type"`   // 1: gợi ý, 0: ngẫu nhiên, 2: tìm kiếm
	Interaction InteractionInfo `json:"interaction"` // Trường interaction mới
}

func GetPostForStudent(c *fiber.Ctx) error {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	var suggestedPostIDs []string
	suggestedPosts, err := flaskClient.CallSuggest(email)
	if err != nil {
		log.Printf("Failed to call Flask suggest API: %v", err)
	} else {
		suggestedPostIDs = suggestedPosts
	}

	var suggestedUUIDs []uuid.UUID
	for _, postID := range suggestedPostIDs {
		if uid, err := uuid.Parse(postID); err == nil {
			suggestedUUIDs = append(suggestedUUIDs, uid)
		}
	}

	var allPosts []models.Post
	if err := database.DB.Db.Preload("Testcase").
		Where("is_deleted = ?", false).
		Find(&allPosts).Error; err != nil {
		log.Printf("Failed to fetch all posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch posts from database",
		})
	}

	var suggestedPostsList []models.Post
	for _, post := range allPosts {
		for _, sugID := range suggestedUUIDs {
			if post.ID == sugID {
				suggestedPostsList = append(suggestedPostsList, post)
				break
			}
		}
	}

	var randomPosts []models.Post
	for _, post := range allPosts {
		isSuggested := false
		for _, sugID := range suggestedUUIDs {
			if post.ID == sugID {
				isSuggested = true
				break
			}
		}
		if !isSuggested {
			randomPosts = append(randomPosts, post)
		}
	}

	numRandom := 3
	if len(randomPosts) < numRandom {
		numRandom = len(randomPosts)
	}
	rand.Shuffle(len(randomPosts), func(i, j int) {
		randomPosts[i], randomPosts[j] = randomPosts[j], randomPosts[i]
	})
	selectedRandomPosts := randomPosts[:numRandom]

	var postIDs []uuid.UUID
	for _, post := range suggestedPostsList {
		postIDs = append(postIDs, post.ID)
	}
	for _, post := range selectedRandomPosts {
		postIDs = append(postIDs, post.ID)
	}

	stats := services.GetPostStats(email, postIDs)
	statsMap := make(map[uuid.UUID]services.PostStats)
	for _, stat := range stats {
		statsMap[stat.PostID] = stat
	}

	var users []models.User
	if err := database.DB.Db.Find(&users).Error; err != nil {
		log.Printf("Failed to fetch users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[user.Mail] = user.LastName + " " + user.FirstName
	}

	var resultPosts []PostWithType
	suggestedCount := len(suggestedPostsList)
	randomCount := len(selectedRandomPosts)
	maxLen := suggestedCount + randomCount

	sugIdx, randIdx := 0, 0
	for i := 0; i < maxLen; i++ {
		var post models.Post
		var postType int
		if i%2 == 0 && sugIdx < suggestedCount {
			post = suggestedPostsList[sugIdx]
			postType = 1 // Gợi ý
			sugIdx++
		} else if randIdx < randomCount {
			post = selectedRandomPosts[randIdx]
			postType = 0 // Ngẫu nhiên
			randIdx++
		} else if sugIdx < suggestedCount {
			post = suggestedPostsList[sugIdx]
			postType = 1 // Gợi ý
			sugIdx++
		} else {
			continue
		}

		stat, exists := statsMap[post.ID]
		if !exists {
			stat = services.PostStats{}
		}
		author := userMap[post.UserMail]
		if author == "" {
			author = "Unknown" // Giá trị mặc định nếu không tìm thấy user
		}

		resultPosts = append(resultPosts, PostWithType{
			Post:     post,
			Author:   author,
			PostType: postType,
			Interaction: InteractionInfo{
				LikeCount:           stat.LikeCount,
				CommentCount:        stat.CommentCount,
				LikeID:              stat.LikeID,
				VerifiedTeacherMail: stat.VerifiedTeacherMail,
				Views:               stat.Views, // Lấy từ PostStats
				Runs:                stat.Runs,  // Lấy từ PostStats
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(resultPosts)
}

func ReadPost(c *fiber.Ctx) error {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	type ReadPostRequest struct {
		PostID   uuid.UUID `json:"post_id"`
		PostType int       `json:"post_type"`
	}
	req := new(ReadPostRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body: " + err.Error(),
		})
	}

	err := database.DB.Db.Transaction(func(tx *gorm.DB) error {
		var post models.Post
		if err := tx.First(&post, "id = ? AND is_deleted = ?", req.PostID, false).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Post not found or has been deleted",
			})
		}

		var user models.User
		if err := tx.First(&user, "mail = ?", email).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		post.Views++
		switch req.PostType {
		case 1:
			post.ViewsBySuggest++
		case 2:
			post.ViewsBySearch++
		}

		user.ReadPosts++
		switch req.PostType {
		case 0:
			user.ReadRandomPosts++
		case 1:
			user.ReadSuggestedPosts++
		case 2:
			user.ReadSearchPosts++
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid post_type value",
			})
		}

		interaction := models.PostInteraction{
			ID:       uuid.New(),
			UserMail: email,
			PostID:   req.PostID,
			PostType: req.PostType,
			Action:   "view",
		}

		if err := tx.Create(&interaction).Error; err != nil {
			return err
		}

		if err := tx.Save(&post).Error; err != nil {
			return err
		}
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if _, ok := err.(*fiber.Error); ok {
			return err
		}
		log.Printf("Failed to update read stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update read stats",
		})
	}

	var user models.User
	database.DB.Db.First(&user, "mail = ?", email)
	return c.Status(fiber.StatusOK).JSON(user)
}

func SearchPosts(c *fiber.Ctx) error {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	var searchPosts []models.Post
	if err := database.DB.Db.Preload("Testcase").
		Where("title LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%").
		Where("is_deleted = ?", false).
		Find(&searchPosts).Error; err != nil {
		log.Printf("Failed to search posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search posts in database",
		})
	}

	// Cập nhật ViewsBySearch và SearchPosts
	err := database.DB.Db.Transaction(func(tx *gorm.DB) error {
		for i := range searchPosts {
			// Tăng Views trong Post
			searchPosts[i].Views++
			searchPosts[i].ViewsBySearch++
			if err := tx.Save(&searchPosts[i]).Error; err != nil {
				return err
			}

			// Ghi vào PostInteraction
			interaction := models.PostInteraction{
				ID:       uuid.New(),
				UserMail: email,
				PostID:   searchPosts[i].ID,
				PostType: 2, // Search
				Action:   "view",
			}
			if err := tx.Create(&interaction).Error; err != nil {
				return err
			}
		}

		// Tăng ReadSearchPosts trong User
		var user models.User
		if err := tx.First(&user, "mail = ?", email).Error; err == nil {
			user.ReadPosts += len(searchPosts)
			user.ReadSearchPosts += len(searchPosts)
			if err := tx.Save(&user).Error; err != nil {
				log.Printf("Failed to update user ReadSearchPosts: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Failed to update search stats: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update search stats",
		})
	}

	var postIDs []uuid.UUID
	for _, post := range searchPosts {
		postIDs = append(postIDs, post.ID)
	}

	stats := services.GetPostStats(email, postIDs)
	statsMap := make(map[uuid.UUID]services.PostStats)
	for _, stat := range stats {
		statsMap[stat.PostID] = stat
	}
	var users []models.User
	if err := database.DB.Db.Find(&users).Error; err != nil {
		log.Printf("Failed to fetch users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}
	userMap := make(map[string]string)
	for _, user := range users {
		userMap[user.Mail] = user.LastName + " " + user.FirstName
	}

	var resultPosts []PostWithType
	for _, post := range searchPosts {
		stat, exists := statsMap[post.ID]
		if !exists {
			stat = services.PostStats{}
		}
		author := userMap[post.UserMail]
		if author == "" {
			author = "Unknown"
		}

		resultPosts = append(resultPosts, PostWithType{
			Post:     post,
			Author:   author,
			PostType: 2, // Tìm kiếm
			Interaction: InteractionInfo{
				LikeCount:           stat.LikeCount,
				CommentCount:        stat.CommentCount,
				LikeID:              stat.LikeID,
				VerifiedTeacherMail: stat.VerifiedTeacherMail,
				Views:               stat.Views, // Lấy từ PostStats
				Runs:                stat.Runs,  // Lấy từ PostStats
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(resultPosts)
}
