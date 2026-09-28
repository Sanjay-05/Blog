package service

import (
	"blog/models"
	"blog/repository"
	"blog/request"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
)

// PostService defines post business logic
type PostService interface {
	CreatePost(userID uuid.UUID, req request.CreatePostRequest) (*models.Post, error)
	GetPostByID(id uuid.UUID) (*models.Post, error)
	GetPostBySlug(slug string) (*models.Post, error)
	GetAllPosts(limit, offset int, status string) ([]models.Post, int64, error)
	GetPostsByCategory(categoryID uuid.UUID, limit, offset int) ([]models.Post, int64, error)
	GetPostsByUser(userID uuid.UUID, limit, offset int) ([]models.Post, int64, error)
	UpdatePost(id uuid.UUID, req request.UpdatePostRequest) (*models.Post, error)
	DeletePost(id uuid.UUID) error
	GetTotalPostsCount() (int64, error)
	GetFirstPostDate() (time.Time, error)
}

type postService struct {
	postRepo     repository.PostRepository
	categoryRepo repository.CategoryRepository
	commentRepo  repository.CommentRepository
	userRepo     repository.UserRepository
	logger       *logrus.Logger
}

// NewPostService creates a new post service
func NewPostService(postRepo repository.PostRepository, categoryRepo repository.CategoryRepository, commentRepo repository.CommentRepository, userRepo repository.UserRepository, logger *logrus.Logger) PostService {
	return &postService{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		commentRepo:  commentRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

// CreatePost creates a new post
func (s *postService) CreatePost(userID uuid.UUID, req request.CreatePostRequest) (*models.Post, error) {
	// Verify category exists
	_, err := s.categoryRepo.GetCategoryByID(req.CategoryID)
	if err != nil {
		s.logger.Warn("Failed to create post - category not found", "category_id", req.CategoryID)
		return nil, ErrForeignKeyViolation
	}

	// Generate slug if not provided
	slugValue := req.Slug
	if slugValue == "" {
		slugValue = slug.Make(req.Title)
	}

	// Check if slug already exists
	existing, _ := s.postRepo.GetPostBySlug(slugValue)
	if existing != nil {
		s.logger.Warn("Failed to create post - slug already exists", "slug", slugValue)
		return nil, ErrUniqueViolation
	}
	statusValue := req.Status
	if statusValue == "" {
		statusValue = "draft"
	}

	post := &models.Post{
		Title:      req.Title,
		Slug:       slugValue,
		Content:    req.Content,
		Status:     statusValue,
		UserID:     userID,
		CategoryID: req.CategoryID,
	}

	if err := s.postRepo.CreatePost(post); err != nil {
		s.logger.Error("Failed to create post", err)
		return nil, err
	}

	s.logger.Info("Post created successfully", "post_id", post.ID, "user_id", userID)
	return s.postRepo.GetPostByID(post.ID)
}

// GetPostByID gets post by ID
func (s *postService) GetPostByID(id uuid.UUID) (*models.Post, error) {
	return s.postRepo.GetPostByID(id)
}

// GetPostBySlug gets post by slug
func (s *postService) GetPostBySlug(slug string) (*models.Post, error) {
	return s.postRepo.GetPostBySlug(slug)
}

// GetAllPosts gets all posts
func (s *postService) GetAllPosts(limit, offset int, status string) ([]models.Post, int64, error) {
	return s.postRepo.GetAllPosts(limit, offset, status)
}

// GetPostsByCategory gets posts by category
func (s *postService) GetPostsByCategory(categoryID uuid.UUID, limit, offset int) ([]models.Post, int64, error) {
	_, err := s.categoryRepo.GetCategoryByID(categoryID)
	if err != nil {
		s.logger.Warn("failed to get post by category - category not found")
		return nil, 0, ErrNotFound
	}
	return s.postRepo.GetPostsByCategory(categoryID, limit, offset)
}

// GetPostsByUser gets posts by user
func (s *postService) GetPostsByUser(userID uuid.UUID, limit, offset int) ([]models.Post, int64, error) {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		s.logger.Warn("failed to get user post - user not found", userID)
		return nil, 0, ErrNotFound
	}
	return s.postRepo.GetPostsByUser(userID, limit, offset)
}

// UpdatePost updates a post
func (s *postService) UpdatePost(id uuid.UUID, req request.UpdatePostRequest) (*models.Post, error) {
	_, err := s.postRepo.GetPostByID(id)
	if err != nil {
		s.logger.Warn("Failed to update post - post not found", "post_id", id)
		return nil, ErrNotFound
	}

	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = *req.Title
	}

	if req.Slug != nil {
		existingPost, err := s.postRepo.GetPostBySlug(*req.Slug)
		if err == nil && existingPost.ID != id {
			s.logger.Warn("Failed to update post - slug already exists", "slug", *req.Slug)
			return nil, ErrUniqueViolation
		}
		updates["slug"] = *req.Slug
	}

	if req.Content != nil {
		updates["content"] = *req.Content
	}

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.CategoryID != nil {
		_, err := s.categoryRepo.GetCategoryByID(*req.CategoryID)
		if err != nil {
			s.logger.Warn("Failed to update post - category not found", "category_id", *req.CategoryID)
			return nil, ErrForeignKeyViolation
		}
		updates["category_id"] = *req.CategoryID
	}

	if err := s.postRepo.UpdatePost(id, updates); err != nil {
		s.logger.Error("Failed to update post", err)
		return nil, err
	}

	s.logger.Info("Post updated successfully", "post_id", id)
	return s.postRepo.GetPostByID(id)
}

// DeletePost deletes a post
func (s *postService) DeletePost(id uuid.UUID) error {
	post, err := s.postRepo.GetPostByID(id)
	if err != nil {
		return ErrNotFound
	}

	if err := s.commentRepo.DeleteCommentsByPost(post.ID); err != nil {
		return err
	}

	updates := map[string]interface{}{
		"status": "deleted",
	}
	if err := s.postRepo.UpdatePost(id, updates); err != nil {
		return err
	}
	return nil
}

// GetTotalPostsCount gets total posts count
func (s *postService) GetTotalPostsCount() (int64, error) {
	return s.postRepo.GetTotalPostsCount()
}

func (s *postService) GetFirstPostDate() (time.Time, error) {
	return s.postRepo.GetFirstPostDate()
}
