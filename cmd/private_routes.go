package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/handlers"
	"github.com/tison2810/be-go-tc/middleware"
)

func privateRoutes(app *fiber.App) {
	// private := app.Group("/api/private", middleware.AuthMiddleware())
	private := app.Group("/", middleware.AuthMiddleware())
	// private.Post("/create", handlers.CreatePost)
	private.Post("/create", handlers.CreatePost)
	private.Post("/confirm/:id", handlers.PostAnyway)
	// private.Put("/post/:id", handlers.UpdatePostFormData)
	private.Get("/posts", handlers.GetAllPosts)
	private.Get("/posts/hot", handlers.GetHotPosts)
	private.Get("/postsID", handlers.GetAllPostsID)
	private.Get("/post/:id", handlers.GetPost)
	// private.Put("/post/:id", handlers.UpdatePost)
	private.Delete("/delete/:id", handlers.DeletePost)
	private.Put("/post/:id/like", handlers.LikePost)
	private.Get("/post/:id/related", handlers.GetRelatedPosts)
	private.Get("/post/:id/comments", handlers.GetPostComment)
	private.Post("/posts/read", handlers.ReadPost)
	private.Post("/posts/search", handlers.SearchPosts)
	private.Get("/checkfile", handlers.CheckFileExist)
	private.Get("/sgposts", handlers.GetPostForStudent)
	private.Post("/verify/:id", handlers.VerifyPost)
	// private.Post("/comment", handlers.CreateComment)
	// private.Put("/comment/:id", handlers.UpdateComment)
	private.Delete("/comment/:id", handlers.DeleteComment)
	private.Post("/comment", handlers.CreateCommentFormData)
	private.Put("/comment/:id", handlers.UpdateCommentFormData)

	private.Get("/jobe/languages", handlers.CheckJobeLanguages)
	private.Put("/jobe/files/:id", handlers.UploadSingleFileToJobeHandler)
	private.Head("/jobe/files/:id", handlers.CheckFile)
	private.Post("/jobe/run", handlers.SubmitRun)

	private.Post("/upload", handlers.UploadTwoFilesHandler)
	private.Get("/runcode/:id", handlers.RunCode)

	private.Get("/user/posts", handlers.GetUserPosts)
	private.Get("/user/likedposts", handlers.GetLikedPosts)
	private.Get("/user/commentposts/:id", handlers.GetPostComment)
	private.Get("/user/comments", handlers.GetUserComments)

	// private.Post("/interactions", handlers.CreateInteraction)
	// private.Get("/interactions", handlers.GetAllInteractions)
	// private.Get("/interactions/:id", handlers.GetInteraction)
	// private.Put("/interactions/:id", handlers.UpdateInteraction)
	// private.Delete("/interactions/:id", handlers.DeleteInteraction)
}
