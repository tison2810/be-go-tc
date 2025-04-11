package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/models"
	"github.com/tison2810/be-go-tc/services"
)

func GetLikedPosts(c *fiber.Ctx) error {
	// Lấy email từ Locals (do AuthMiddleware cung cấp)
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Truy vấn các bài post đã like
	var interactions []models.Interaction
	if err := database.DB.Db.Where("user_mail = ? AND is_like = ?", userMail, true).Find(&interactions).Error; err != nil {
		log.Printf("Failed to fetch liked posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch liked posts",
		})
	}

	// Nếu không có bài post nào được like
	if len(interactions) == 0 {
		return c.Status(fiber.StatusOK).JSON([]PostWithType{})
	}

	// Lấy danh sách post_id từ interactions
	var postIDs []uuid.UUID
	for _, interaction := range interactions {
		postIDs = append(postIDs, interaction.PostID)
	}

	// Truy vấn các bài post tương ứng
	var posts []models.Post
	if err := database.DB.Db.Where("id IN ? AND is_deleted = ?", postIDs, false).
		Preload("Testcase").Find(&posts).Error; err != nil {
		log.Printf("Failed to fetch posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch posts",
		})
	}

	// Lấy thông tin tương tác từ GetPostStats
	stats := services.GetPostStats(userMail, postIDs)
	statsMap := make(map[uuid.UUID]models.PostStats)
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
			stat = models.PostStats{}
		}

		// Tạo Author từ userMap
		author := userMap[post.UserMail]
		if author == "" {
			author = "Unknown"
		}

		// Mặc định PostType là 0 (ngẫu nhiên)
		postType := 0

		resultPosts = append(resultPosts, PostWithType{
			Post:     post,
			Author:   author,
			PostType: postType,
			Interaction: models.InteractionInfo{
				LikeCount:           stat.LikeCount,
				CommentCount:        stat.CommentCount,
				LikeID:              stat.LikeID,
				VerifiedTeacherMail: stat.VerifiedTeacherMail,
				Views:               stat.Views,
				Runs:                stat.Runs,
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(resultPosts)
}

func GetUserPosts(c *fiber.Ctx) error {
	// Lấy email từ Locals
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Truy vấn các bài post của người dùng
	var posts []models.Post
	if err := database.DB.Db.Where("user_mail = ? AND is_deleted = ?", userMail, false).
		Preload("Testcase").Find(&posts).Error; err != nil {
		log.Printf("Failed to fetch user posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user posts",
		})
	}

	// Nếu không có bài post nào
	if len(posts) == 0 {
		return c.Status(fiber.StatusOK).JSON([]PostWithType{})
	}

	// Lấy danh sách post_id
	var postIDs []uuid.UUID
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	// Lấy thông tin tương tác từ GetPostStats
	stats := services.GetPostStats(userMail, postIDs)
	statsMap := make(map[uuid.UUID]models.PostStats)
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
			stat = models.PostStats{}
		}

		// Tạo Author từ userMap
		author := userMap[post.UserMail]
		if author == "" {
			author = "Unknown"
		}

		// Mặc định PostType là 0 (ngẫu nhiên)
		postType := 0

		resultPosts = append(resultPosts, PostWithType{
			Post:     post,
			Author:   author,
			PostType: postType,
			Interaction: models.InteractionInfo{
				LikeCount:           stat.LikeCount,
				CommentCount:        stat.CommentCount,
				LikeID:              stat.LikeID,
				VerifiedTeacherMail: stat.VerifiedTeacherMail,
				Views:               stat.Views,
				Runs:                stat.Runs,
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(resultPosts)
}

func GetCommentedPosts(c *fiber.Ctx) error {
	// Lấy email từ Locals
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Truy vấn các comment của người dùng
	var comments []models.Comment
	if err := database.DB.Db.Where("user_mail = ? AND is_deleted = ?", userMail, false).
		Find(&comments).Error; err != nil {
		log.Printf("Failed to fetch user comments: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user comments",
		})
	}

	// Nếu không có comment nào
	if len(comments) == 0 {
		return c.Status(fiber.StatusOK).JSON([]PostWithType{})
	}

	// Lấy danh sách post_id (loại bỏ trùng lặp)
	postIDSet := make(map[uuid.UUID]bool)
	var postIDs []uuid.UUID
	for _, comment := range comments {
		if !postIDSet[comment.PostID] {
			postIDSet[comment.PostID] = true
			postIDs = append(postIDs, comment.PostID)
		}
	}

	// Truy vấn các bài post tương ứng
	var posts []models.Post
	if err := database.DB.Db.Where("id IN ? AND is_deleted = ?", postIDs, false).
		Preload("Testcase").Find(&posts).Error; err != nil {
		log.Printf("Failed to fetch commented posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch commented posts",
		})
	}

	// Lấy thông tin tương tác từ GetPostStats
	stats := services.GetPostStats(userMail, postIDs)
	statsMap := make(map[uuid.UUID]models.PostStats)
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
			stat = models.PostStats{}
		}

		// Tạo Author từ userMap
		author := userMap[post.UserMail]
		if author == "" {
			author = "Unknown"
		}

		// Mặc định PostType là 0 (ngẫu nhiên)
		postType := 0

		resultPosts = append(resultPosts, PostWithType{
			Post:     post,
			Author:   author,
			PostType: postType,
			Interaction: models.InteractionInfo{
				LikeCount:           stat.LikeCount,
				CommentCount:        stat.CommentCount,
				LikeID:              stat.LikeID,
				VerifiedTeacherMail: stat.VerifiedTeacherMail,
				Views:               stat.Views,
				Runs:                stat.Runs,
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(resultPosts)
}

func GetUserComments(c *fiber.Ctx) error {
	// Lấy email từ Locals
	userMail, ok := c.Locals("email").(string)
	if !ok || userMail == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User email not found in context",
		})
	}

	// Truy vấn các comment của người dùng
	var comments []models.Comment
	if err := database.DB.Db.Where("user_mail = ? AND is_deleted = ?", userMail, false).
		Find(&comments).Error; err != nil {
		log.Printf("Failed to fetch user comments: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user comments",
		})
	}

	// Nếu không có comment nào
	if len(comments) == 0 {
		return c.Status(fiber.StatusOK).JSON([]fiber.Map{})
	}

	// Lấy danh sách post_id (loại bỏ trùng lặp)
	postIDSet := make(map[uuid.UUID]bool)
	var postIDs []uuid.UUID
	for _, comment := range comments {
		if !postIDSet[comment.PostID] {
			postIDSet[comment.PostID] = true
			postIDs = append(postIDs, comment.PostID)
		}
	}

	// Truy vấn các bài post tương ứng
	var posts []models.Post
	if err := database.DB.Db.Where("id IN ? AND is_deleted = ?", postIDs, false).
		Preload("Testcase").Find(&posts).Error; err != nil {
		log.Printf("Failed to fetch posts: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch posts",
		})
	}

	// Lấy thông tin tương tác từ GetPostStats
	stats := services.GetPostStats(userMail, postIDs)
	statsMap := make(map[uuid.UUID]models.PostStats)
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

	// Tạo danh sách kết quả: mỗi comment kèm thông tin bài post
	var resultComments []fiber.Map
	for _, comment := range comments {
		// Tìm bài post tương ứng
		var post models.Post
		for _, p := range posts {
			if p.ID == comment.PostID {
				post = p
				break
			}
		}

		// Nếu không tìm thấy bài post (có thể bài post đã bị xóa), bỏ qua comment này
		if post.ID == uuid.Nil {
			continue
		}

		// Lấy PostStats
		stat, exists := statsMap[post.ID]
		if !exists {
			stat = models.PostStats{}
		}

		// Tạo Author từ userMap
		author := userMap[post.UserMail]
		if author == "" {
			author = "Unknown"
		}

		// Mặc định PostType là 0 (ngẫu nhiên)
		postType := 0

		// Tạo PostWithType cho bài post
		postWithType := PostWithType{
			Post:     post,
			Author:   author,
			PostType: postType,
			Interaction: models.InteractionInfo{
				LikeCount:           stat.LikeCount,
				CommentCount:        stat.CommentCount,
				LikeID:              stat.LikeID,
				VerifiedTeacherMail: stat.VerifiedTeacherMail,
				Views:               stat.Views,
				Runs:                stat.Runs,
			},
		}

		// Thêm comment và bài post vào kết quả
		resultComments = append(resultComments, fiber.Map{
			"comment": comment,
			"post":    postWithType,
		})
	}

	return c.Status(fiber.StatusOK).JSON(resultComments)
}
