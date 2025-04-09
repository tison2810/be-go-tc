package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/models"
	"gorm.io/gorm"
)

// VerifyPost cho phép giảng viên xác minh một bài post
func VerifyPost(c *fiber.Ctx) error {
	// Lấy email và role từ Locals (do AuthMiddleware cung cấp)
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	role, ok := c.Locals("role").(string)
	if !ok || role == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User role not found in context",
		})
	}

	// Kiểm tra role phải là "teacher"
	if role != "teacher" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only teachers can verify posts",
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

	// Dùng transaction để đảm bảo tính nhất quán
	err = database.DB.Db.Transaction(func(tx *gorm.DB) error {
		// Kiểm tra post tồn tại và chưa bị xóa
		var post models.Post
		if err := tx.First(&post, "id = ? AND is_deleted = ?", postID, false).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Post not found or has been deleted",
				})
			}
			return err
		}

		// Kiểm tra xem post đã được verify chưa (dựa trên TeacherVerifyPost)
		var existingVerify models.TeacherVerifyPost
		if err := tx.First(&existingVerify, "post_id = ?", postID).Error; err == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Post has already been verified by a teacher",
			})
		} else if err != gorm.ErrRecordNotFound {
			return err
		}

		// Thêm bản ghi vào TeacherVerifyPost
		teacherVerify := models.TeacherVerifyPost{
			PostID:      postID,
			TeacherMail: email,
		}
		if err := tx.Create(&teacherVerify).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// Nếu lỗi đã được trả về trong transaction (StatusNotFound, StatusBadRequest), không ghi đè
		if _, ok := err.(*fiber.Error); ok {
			return err
		}
		log.Printf("Failed to verify post: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to verify post",
		})
	}

	// Trả về thông báo thành công
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Post verified successfully",
		"post_id":      postID,
		"teacher_mail": email,
	})
}
