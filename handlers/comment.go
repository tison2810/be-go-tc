package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/models"

	"github.com/tison2810/be-go-tc/database"
)

func CreateComment(c *fiber.Ctx) error {
	comment := new(models.Comment)
	if err := c.BodyParser(comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	comment.ID = uuid.New()
	comment.CreatedAt = time.Now()

	database.DB.Db.Create(&comment)
	return c.Status(fiber.StatusCreated).JSON(comment)
}

func GetPostComment(c *fiber.Ctx) error {
	id := c.Params("id")
	var comments []models.Comment
	database.DB.Db.Where("post_id = ?", id).Find(&comments)
	return c.Status(fiber.StatusOK).JSON(comments)
}

func GetAllComments(c *fiber.Ctx) error {
	var comments []models.Comment
	database.DB.Db.Preload("Post").Find(&comments)
	return c.Status(fiber.StatusOK).JSON(comments)
}

func UpdateComment(c *fiber.Ctx) error {
	id := c.Params("id")
	comment := new(models.Comment)

	database.DB.Db.Where("id = ?", id).First(&comment)
	if comment.ID == uuid.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	if err := c.BodyParser(comment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := database.DB.Db.Save(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update post",
		})
	}

	return c.Status(fiber.StatusOK).JSON(comment)
}

func DeleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	comment := new(models.Comment)
	database.DB.Db.Where("id = ?", id).First(&comment)
	if comment.ID == uuid.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Comment not found",
		})
	}

	database.DB.Db.Delete(&comment)
	return c.Status(fiber.StatusNoContent).Send(nil)
}
