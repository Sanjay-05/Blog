package router

import (
	"blog/config"
	"blog/handler"
	"blog/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

// InitializeRouter initializes all routes
func InitializeRouter(
	app *fiber.App,
	cfg *config.Configuration,
	userHandler handler.UserHandler,
	categoryHandler handler.CategoryHandler,
	postHandler handler.PostHandler,
	commentHandler handler.CommentHandler,
	logger *logrus.Logger,
) {
	// Public routes
	api := app.Group("/api/v1")

	{
		// Auth routes (public)
		auth := api.Group("/auth")
		auth.Post("/register", userHandler.Register)
		auth.Post("/login", userHandler.Login)
		auth.Post("/forget-password", userHandler.ForgetPasswordHandler)
		auth.Post("/reset-password", userHandler.ResetPassword)
	}

	// Public blog routes
	public := api.Group("/public")
	{
		posts := public.Group("/posts")
		posts.Get("", postHandler.GetAllPosts)
		posts.Get("/:id", postHandler.GetPost)
		posts.Get("/comments/:post_id", commentHandler.GetCommentsByPost)
	}

	public.Get("/categories/with-count", categoryHandler.GetAllCategoriesWithCount)
	public.Get("/categories", categoryHandler.GetAllCategories)
	public.Get("/categories/:id", categoryHandler.GetCategory)

	// Protected routes (authenticated users)
	protected := api.Group("")
	protected.Use(middleware.JWTMiddleware(cfg, logger))

	// User routes
	user := protected.Group("/users")
	user.Get("/profile", userHandler.GetProfile)
	user.Put("/modify", userHandler.UpdateUser)
	user.Delete("/remove", userHandler.DeleteUser)

	// Post routes
	posts := protected.Group("/posts")
	posts.Get("/user", postHandler.GetPostsByUser)

	// Comment routes
	comments := protected.Group("/comments")
	comments.Post("/", commentHandler.CreateComment)
	comments.Get("/user", commentHandler.GetCommentsByUser)
	comments.Get("/:id", commentHandler.GetComment)
	comments.Put("/modify", commentHandler.UpdateComment)
	comments.Delete("/:id", commentHandler.DeleteComment)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(middleware.AdminMiddleware())

	// Admin user routes
	{
		adminUsers := admin.Group("/users")
		adminUsers.Get("/", userHandler.GetAllUsers)
		adminUsers.Patch("/switch", userHandler.SwitchUserRole)
		adminUsers.Delete("/remove/user", userHandler.DeleteUser)
	}

	//Admin Posts routes
	{
		adminPosts := admin.Group("/posts")
		adminPosts.Post("/", postHandler.CreatePost)
		adminPosts.Put("/modify/post", postHandler.UpdatePost)
		adminPosts.Delete("/:id", postHandler.DeletePost)
	}

	// Admin category routes
	{
		adminCategories := admin.Group("/categories")
		adminCategories.Post("/", categoryHandler.CreateCategory)
		adminCategories.Put("/modify", categoryHandler.UpdateCategory)
		adminCategories.Delete("/:id", categoryHandler.DeleteCategory)
	}

	// Admin stats
	adminStats := admin.Group("/stats")
	adminStats.Get("/overview", postHandler.GetAdminStats)

	// Admin comment routes
	adminComments := admin.Group("/comments")
	adminComments.Delete("/:id", commentHandler.AdminDeleteComment)
}
