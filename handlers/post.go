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
	"github.com/tison2810/be-go-tc/utils"
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
	post.CreatedAt = time.Now()
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
		fmt.Print(post.Trace)
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
	post.CreatedAt = time.Now()

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
		fmt.Print(post.Trace)
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

func GetAllPosts(c *fiber.Ctx) error {
	var posts []models.Post
	database.DB.Db.Preload("Testcase").Find(&posts)
	return c.Status(fiber.StatusOK).JSON(posts)
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

	database.DB.Db.Delete(&post)
	return c.Status(fiber.StatusNoContent).Send(nil)
}

func CreateInteraction(c *fiber.Ctx) error {
	// Lấy email từ Locals (do AuthMiddleware cung cấp)
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Parse request body
	interaction := new(models.Interaction)
	if err := c.BodyParser(interaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	// Thiết lập các giá trị mặc định
	interaction.ID = uuid.New()
	interaction.UserMail = userMail
	interaction.CreatedAt = time.Now()

	// Validate Type
	if interaction.Type != "Like" && interaction.Type != "Rating" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Type must be 'Like' or 'Rating'",
		})
	}

	// Validate Rating nếu Type là "Rating"
	if interaction.Type == "Rating" && (interaction.Rating < 0 || interaction.Rating > 5) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Rating must be between 0 and 5",
		})
	}

	// Lưu vào database
	if err := database.DB.Db.Create(&interaction).Error; err != nil {
		log.Printf("Failed to save interaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save interaction: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(interaction)
}

func LikePost(c *fiber.Ctx) error {
	// Lấy email từ Locals (do AuthMiddleware cung cấp)
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Parse request body
	interaction := new(models.Interaction)
	if err := c.BodyParser(interaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	// Thiết lập các giá trị mặc định
	interaction.ID = uuid.New()
	interaction.UserMail = userMail
	interaction.CreatedAt = time.Now()
	interaction.Type = "Like"
	interaction.Rating = 0
	interaction.IsLike = true

	// Lưu vào database
	if err := database.DB.Db.Create(&interaction).Error; err != nil {
		log.Printf("Failed to save interaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save interaction: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(interaction)
}

func RatingPost(c *fiber.Ctx) error {
	// Lấy email từ Locals (do AuthMiddleware cung cấp)
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Parse request body
	interaction := new(models.Interaction)
	if err := c.BodyParser(interaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	// Thiết lập các giá trị mặc định
	interaction.ID = uuid.New()
	interaction.UserMail = userMail
	interaction.CreatedAt = time.Now()
	interaction.Type = "Rating"
	interaction.IsLike = false

	// Lưu vào database
	if err := database.DB.Db.Create(&interaction).Error; err != nil {
		log.Printf("Failed to save interaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save interaction: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(interaction)
}
func GetAllInteractions(c *fiber.Ctx) error {
	var interactions []models.Interaction
	if err := database.DB.Db.Find(&interactions).Error; err != nil {
		log.Printf("Failed to fetch interactions: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch interactions: " + err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(interactions)
}

func GetInteraction(c *fiber.Ctx) error {
	id := c.Params("id")
	interaction := new(models.Interaction)
	if err := database.DB.Db.Where("id = ?", id).First(&interaction).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Interaction not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(interaction)
}

func UpdateInteraction(c *fiber.Ctx) error {
	id := c.Params("id")
	interaction := new(models.Interaction)
	if err := database.DB.Db.Where("id = ?", id).First(&interaction).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Interaction not found",
		})
	}

	// Parse request body để cập nhật
	if err := c.BodyParser(interaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	// Validate Type
	if interaction.Type != "Like" && interaction.Type != "Rating" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Type must be 'Like' or 'Rating'",
		})
	}

	// Validate Rating nếu Type là "Rating"
	if interaction.Type == "Rating" && (interaction.Rating < 0 || interaction.Rating > 5) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Rating must be between 0 and 5",
		})
	}

	// Cập nhật trong database
	if err := database.DB.Db.Save(&interaction).Error; err != nil {
		log.Printf("Failed to update interaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update interaction: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(interaction)
}

func DeleteInteraction(c *fiber.Ctx) error {
	id := c.Params("id")
	interaction := new(models.Interaction)
	if err := database.DB.Db.Where("id = ?", id).First(&interaction).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Interaction not found",
		})
	}

	if err := database.DB.Db.Delete(&interaction).Error; err != nil {
		log.Printf("Failed to delete interaction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete interaction: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func GetUserLikeStatus(c *fiber.Ctx) error {
	// Lấy email từ Locals (do AuthMiddleware cung cấp)
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Lấy post_id từ param
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

	// Kiểm tra xem người dùng đã Like bài post này chưa
	var interaction models.Interaction
	result := database.DB.Db.Where("user_mail = ? AND post_id = ? AND type = ? AND is_like = ?",
		userMail, postID, "Like", true).First(&interaction)

	// Nếu không tìm thấy bản ghi, tức là chưa Like
	if result.Error != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"is_liked": false,
		})
	}

	// Nếu tìm thấy, trả về is_liked = true
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"is_liked":       true,
		"interaction_id": interaction.ID,
	})
}

type PostWithClass struct {
	models.Post
	Classed int `json:"classed"` // 1: gợi ý, 0: ngẫu nhiên
}

func GetPostForStudent(c *fiber.Ctx) error {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Lấy danh sách bài post gợi ý từ Flask
	var suggestedPostIDs []string
	suggestedPosts, err := flaskClient.CallSuggest(email)
	if err != nil {
		log.Printf("Failed to call Flask suggest API: %v", err)
	} else {
		suggestedPostIDs = suggestedPosts
		log.Printf("Suggested post IDs: %v", suggestedPostIDs)
	}

	// Chuyển suggestedPostIDs thành uuid.UUID
	var suggestedUUIDs []uuid.UUID
	for _, postID := range suggestedPostIDs {
		if uid, err := uuid.Parse(postID); err == nil {
			suggestedUUIDs = append(suggestedUUIDs, uid)
		}
	}

	// Lấy tất cả bài post từ database với Testcase
	var allPosts []models.Post
	if err := database.DB.Db.Preload("Testcase").Find(&allPosts).Error; err != nil {
		log.Printf("Failed to fetch all posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch posts from database",
		})
	}

	// Tạo danh sách bài post gợi ý
	var suggestedPostsList []models.Post
	for _, post := range allPosts {
		for _, sugID := range suggestedUUIDs {
			if post.ID == sugID {
				suggestedPostsList = append(suggestedPostsList, post)
				break
			}
		}
	}

	// Lọc ra các bài post không trùng với gợi ý
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

	// Chọn ngẫu nhiên 3 bài từ randomPosts (hoặc ít hơn nếu không đủ)
	numRandom := 3
	if len(randomPosts) < numRandom {
		numRandom = len(randomPosts)
	}
	rand.Shuffle(len(randomPosts), func(i, j int) {
		randomPosts[i], randomPosts[j] = randomPosts[j], randomPosts[i]
	})
	selectedRandomPosts := randomPosts[:numRandom]

	// Tạo danh sách kết quả xen kẽ
	var resultPosts []PostWithClass
	suggestedCount := len(suggestedPostsList)
	randomCount := len(selectedRandomPosts)
	maxLen := suggestedCount + randomCount

	// Xen kẽ bài gợi ý và bài ngẫu nhiên
	sugIdx, randIdx := 0, 0
	for i := 0; i < maxLen; i++ {
		if i%2 == 0 && sugIdx < suggestedCount {
			// Thêm bài gợi ý với suggested = 1
			resultPosts = append(resultPosts, PostWithClass{
				Post:    suggestedPostsList[sugIdx],
				Classed: 1,
			})
			sugIdx++
		} else if randIdx < randomCount {
			// Thêm bài ngẫu nhiên với suggested = 0
			resultPosts = append(resultPosts, PostWithClass{
				Post:    selectedRandomPosts[randIdx],
				Classed: 0,
			})
			randIdx++
		} else if sugIdx < suggestedCount {
			// Nếu hết bài ngẫu nhiên, thêm bài gợi ý còn lại
			resultPosts = append(resultPosts, PostWithClass{
				Post:    suggestedPostsList[sugIdx],
				Classed: 1,
			})
			sugIdx++
		}
	}

	// Trả về danh sách bài post với định dạng giống GetAllPosts
	return c.Status(fiber.StatusOK).JSON(resultPosts)
}

func ReadPost(c *fiber.Ctx) error {
	// Lấy email từ context
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Lấy post_id và classed từ request body
	type ReadPostRequest struct {
		PostID  uuid.UUID `json:"post_id"`
		Classed int       `json:"classed"`
	}
	req := new(ReadPostRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body: " + err.Error(),
		})
	}

	// Kiểm tra post tồn tại
	var post models.Post
	if err := database.DB.Db.First(&post, "id = ?", req.PostID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Post not found",
		})
	}

	// Lấy user từ database
	var user models.User
	if err := database.DB.Db.First(&user, "mail = ?", email).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Tăng ReadPosts
	user.ReadPosts++

	// Tăng các trường dựa trên classed
	switch req.Classed {
	case 0: // Ngẫu nhiên
		user.ReadRandomPosts++
	case 1: // Gợi ý
		user.ReadSuggestedPosts++
	case 2: // Tìm kiếm
		user.ReadSearchPosts++
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid classed value",
		})
	}

	// Cập nhật user trong database
	if err := database.DB.Db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user: " + err.Error(),
		})
	}

	// Trả về thông tin user sau khi cập nhật
	return c.Status(fiber.StatusOK).JSON(user)
}

func SearchPosts(c *fiber.Ctx) error {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Lấy query từ request (ví dụ: ?query=hashing)
	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Search query is required",
		})
	}

	// Tìm kiếm bài post trong database
	var searchPosts []models.Post
	if err := database.DB.Db.Preload("Testcase").
		Where("title LIKE ? OR description LIKE ?", "%"+query+"%", "%"+query+"%").
		Find(&searchPosts).Error; err != nil {
		log.Printf("Failed to search posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search posts in database",
		})
	}

	// Tạo danh sách kết quả với classed = 2
	var resultPosts []PostWithClass
	for _, post := range searchPosts {
		resultPosts = append(resultPosts, PostWithClass{
			Post:    post,
			Classed: 2,
		})
	}

	// Cập nhật SearchPosts trong user
	var user models.User
	if err := database.DB.Db.First(&user, "mail = ?", email).Error; err != nil {
		log.Printf("User not found: %v", err)
	} else {
		user.SearchPosts++
		if err := database.DB.Db.Save(&user).Error; err != nil {
			log.Printf("Failed to update user SearchPosts: %v", err)
		}
	}

	// Trả về danh sách bài post tìm kiếm
	return c.Status(fiber.StatusOK).JSON(resultPosts)
}
