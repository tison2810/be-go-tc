package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/models"

	"github.com/tison2810/be-go-tc/database"
)

func CreatePost(c *fiber.Ctx) error {
	post := new(models.Post)

	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse JSON: " + err.Error(),
		})
	}

	post.ID = uuid.New()
	post.CreatedAt = time.Now()

	if post.Testcase != nil {
		post.Testcase.PostID = post.ID
	}

	if err := database.DB.Db.Create(&post).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save post and testcase: " + err.Error(),
		})
	}

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
