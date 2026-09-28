package main

import (
	"blog/config"
	"blog/driver"
	"blog/handler"
	"blog/repository"
	"blog/router"
	"blog/service"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	db, err := driver.ConnectDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	// Initialize services with logger
	userService := service.NewUserService(userRepo, logger)
	categoryService := service.NewCategoryService(categoryRepo, logger)
	postService := service.NewPostService(postRepo, categoryRepo, commentRepo, userRepo, logger)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo, logger)

	// Initialize handlers with logger
	userHandler := handler.NewUserHandler(userService, cfg, logger)
	categoryHandler := handler.NewCategoryHandler(categoryService, logger)
	postHandler := handler.NewPostHandler(postService, commentService, userService, logger)
	commentHandler := handler.NewCommentHandler(commentService, logger)

	// Initialize Fiber app
	app := fiber.New()

	// Initialize routes
	router.InitializeRouter(app, cfg, userHandler, categoryHandler, postHandler, commentHandler,logger)

	// Log all registered routes
	logger.Info("Blog API started successfully")
	fmt.Printf("Server is running on port %s\n", cfg.ServerPort)

	// Start server
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		logger.Info("Server error", err)
		log.Fatal(err)
	}
}
