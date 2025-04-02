package handlers

import (
	"fmt"
	"log"
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

func CreatePost(c *fiber.Ctx) error {
	post := new(models.Post)

	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	post.ID = uuid.New()
	post.CreatedAt = time.Now()
	post.Subject = "DSA"

	if post.Testcase != nil {
		post.Testcase.PostID = post.ID
	}

	if err := database.DB.Db.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save post and testcase: " + err.Error(),
		})
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
