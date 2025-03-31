package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/handlers"
	"github.com/tison2810/be-go-tc/middleware"
)

func privateRoutes(app *fiber.App) {
	private := app.Group("/", middleware.AuthMiddleware())
	private.Post("/create", handlers.CreatePost)
	private.Get("/posts", handlers.GetAllPosts)
	private.Get("/postsID", handlers.GetAllPostsID)
	private.Get("/post/:id", handlers.GetPost)
	private.Put("/post/:id", handlers.UpdatePost)
	private.Delete("/delete/:id", handlers.DeletePost)

	private.Post("/comment", handlers.CreateComment)
	private.Get("/post/:id/comments", handlers.GetPostComment)
	private.Put("/comment/:id", handlers.UpdateComment)
	private.Delete("/comment/:id", handlers.DeleteComment)

	private.Get("/jobe/languages", handlers.CheckJobeLanguages)
	private.Put("/jobe/files", handlers.UploadFileToJobeHandler)
	private.Head("/jobe/files/:id", handlers.CheckFile)
	private.Post("/jobe/run", handlers.SubmitRun)
}
