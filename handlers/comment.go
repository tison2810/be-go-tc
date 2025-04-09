package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/models"

	"github.com/tison2810/be-go-tc/database"
)

func CreateComment(c *fiber.Ctx) error {
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	comment := new(models.Comment)
	if err := c.BodyParser(comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	comment.ID = uuid.New()
	comment.UserMail = userMail
	comment.CreatedAt = time.Now()
	comment.IsDeleted = false

	if comment.PostID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "PostID is required",
		})
	}

	if err := database.DB.Db.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save comment: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(comment)
}

func GetPostComment(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Post ID is required",
		})
	}

	var comments []models.Comment
	if err := database.DB.Db.Where("post_id = ? AND is_deleted = ?", id, false).Find(&comments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch comments: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(comments)
}

func GetAllComments(c *fiber.Ctx) error {
	var comments []models.Comment
	database.DB.Db.Preload("Post").Find(&comments)
	return c.Status(fiber.StatusOK).JSON(comments)
}

func UpdateComment(c *fiber.Ctx) error {
	// Lấy email từ Locals
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Comment ID is required",
		})
	}

	// Tìm comment
	comment := new(models.Comment)
	if err := database.DB.Db.Where("id = ? AND is_deleted = ?", id, false).First(&comment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found or already deleted",
		})
	}

	// Kiểm tra quyền chỉnh sửa
	if comment.UserMail != userMail {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to update this comment",
		})
	}

	// Parse request body để cập nhật Content
	type UpdateRequest struct {
		Content string `json:"content"`
	}
	req := new(UpdateRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	// Cập nhật Content
	comment.Content = req.Content

	// Lưu vào database
	if err := database.DB.Db.Save(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update comment: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(comment)
}

func DeleteComment(c *fiber.Ctx) error {
	// Lấy email từ Locals
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Comment ID is required",
		})
	}

	// Tìm comment
	comment := new(models.Comment)
	if err := database.DB.Db.Where("id = ? AND is_deleted = ?", id, false).First(&comment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found or already deleted",
		})
	}

	// Kiểm tra quyền xóa
	if comment.UserMail != userMail {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are not authorized to delete this comment",
		})
	}

	// Đánh dấu xóa mềm
	comment.IsDeleted = true
	if err := database.DB.Db.Save(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete comment: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func GetCommentCount(postID uuid.UUID) int64 {
	var commentCount int64
	database.DB.Db.Model(&models.Interaction{}).
		Where("post_id = ? AND type = ?", postID, "Comment").
		Count(&commentCount)
	return commentCount
}
