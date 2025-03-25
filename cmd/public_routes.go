package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/handlers"
)

func publicRoutes(app *fiber.App) {
	app.Post("/create", handlers.CreatePost)
	app.Get("/posts", handlers.GetAllPosts)
	app.Get("/postsID", handlers.GetAllPostsID)
	app.Get("/post/:id", handlers.GetPost)
	app.Put("/post/:id", handlers.UpdatePost)
	app.Delete("/delete/:id", handlers.DeletePost)

	app.Post("/comment", handlers.CreateComment)
	app.Get("/post/:id/comments", handlers.GetPostComment)
	app.Put("/comment/:id", handlers.UpdateComment)
	app.Delete("/comment/:id", handlers.DeleteComment)

	app.Get("/jobe/languages", handlers.CheckJobeLanguages)
	app.Put("/jobe/files", handlers.UploadFileToJobeHandler)
	app.Head("/jobe/files/:id", handlers.CheckFile)
	app.Post("/jobe/run", handlers.SubmitRun)

	app.Post("/auth/google", handlers.GoogleAuthHandler)
}
