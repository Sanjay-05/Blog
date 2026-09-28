package handler

import (
	"blog/helper"
	"blog/models"
	"blog/request"
	"blog/response"
	"blog/service"
	"blog/utility"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// PostHandler handles post-related requests
type PostHandler interface {
	CreatePost(c fiber.Ctx) error
	GetPost(c fiber.Ctx) error
	GetAllPosts(c fiber.Ctx) error
	GetPostsByUser(c fiber.Ctx) error
	UpdatePost(c fiber.Ctx) error
	DeletePost(c fiber.Ctx) error
	GetAdminStats(c fiber.Ctx) error
}

type postHandler struct {
	postService    service.PostService
	commentService service.CommentService
	userService    service.UserService
	logger         *logrus.Logger
}

// NewPostHandler creates a new post handler
func NewPostHandler(ps service.PostService, cs service.CommentService, us service.UserService, logger *logrus.Logger) PostHandler {
	return &postHandler{
		postService:    ps,
		commentService: cs,
		userService:    us,
		logger:         logger,
	}
}

// CreatePost handles post creation (authenticated user)
func (h *postHandler) CreatePost(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to CreatePost")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	var req request.CreatePostRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to bind create post request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		h.logger.Error("Validation failed for create post", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	post, err := h.postService.CreatePost(userID, req)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to create post", err)
		return utility.SendErrorResponse(c, statusCode, "failed to create post", err)
	}

	h.logger.Info("Post created successfully", "post_id", post.ID)
	return utility.SendSuccessResponse(c, http.StatusCreated, "Post created successfully", nil)
}

// GetPost handles getting a single post
func (h *postHandler) GetPost(c fiber.Ctx) error {
	postID := c.Params("id")
	id, err := uuid.Parse(postID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid post id", err)
	}

	post, err := h.postService.GetPostByID(id)
	if err != nil {
		h.logger.Error("Failed to get post", err)
		return utility.SendErrorResponse(c, http.StatusNotFound, "failed to get post", err)
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "Post retrieved successfully", helper.PostDetailResponse(post))
}

// GetAllPosts handles getting all posts, with optional filters for slug and category
func (h *postHandler) GetAllPosts(c fiber.Ctx) error {
	limit, offset := utility.GetPagination(c)
	status := c.Query("status", "published")
	slug := c.Query("slug")
	categoryIDStr := c.Query("category_id")

	var posts []models.Post
	var total int64
	var err error

	switch {
	case slug != "":
		var post *models.Post
		post, err = h.postService.GetPostBySlug(slug)
		if err != nil {
			h.logger.Error("Failed to get post by slug", err)
			statusCode := utility.GetStatusCodeFromError(err)
			return utility.SendErrorResponse(c, statusCode, "post not found", err)
		}
		posts = []models.Post{*post}
		total = 1

	case categoryIDStr != "":
		categoryID, parseErr := uuid.Parse(categoryIDStr)
		if parseErr != nil {
			return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid category id", parseErr)
		}
		posts, total, err = h.postService.GetPostsByCategory(categoryID, limit, offset)
		if err != nil {
			statusCode := utility.GetStatusCodeFromError(err)
			h.logger.Error("Failed to get posts by category", err)
			return utility.SendErrorResponse(c, statusCode, "failed to get posts", err)
		}

	default:
		posts, total, err = h.postService.GetAllPosts(limit, offset, status)
		if err != nil {
			h.logger.Error("Failed to get posts", err)
			statusCode := utility.GetStatusCodeFromError(err)
			return utility.SendErrorResponse(c, statusCode, "failed to get posts", err)
		}
	}

	if len(posts) == 0 {
		return utility.SendSuccessResponse(c, http.StatusOK, "No posts found", map[string]any{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"posts":  []interface{}{},
		})
	}

	var postResponses []response.PostResponse
	for i := range posts {
		postResponses = append(postResponses, helper.PostToResponse(&posts[i]))
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "Posts retrieved successfully", map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"posts":  postResponses,
	})
}

// GetPostsByUser handles getting posts by user
func (h *postHandler) GetPostsByUser(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to get posts")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	limit, offset := utility.GetPagination(c)

	posts, total, err := h.postService.GetPostsByUser(userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get posts by user", err)
		statusCode := utility.GetStatusCodeFromError(err)
		return utility.SendErrorResponse(c, statusCode, "failed to get posts", err)
	}

	if len(posts) == 0 {
		return utility.SendSuccessResponse(c, http.StatusOK, "No posts found", map[string]any{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"posts":  []interface{}{},
		})
	}

	var postResponses []response.PostResponse
	for i := range posts {
		postResponses = append(postResponses, helper.PostToResponse(&posts[i]))
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "Posts retrieved successfully", map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"posts":  postResponses,
	})
}

// UpdatePost handles post update
func (h *postHandler) UpdatePost(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to UpdatePost")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	var req request.UpdatePostRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to bind update post request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	post, err := h.postService.GetPostByID(req.PostID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusNotFound, "failed to get post", err)
	}

	role, _ := c.Locals("role").(string)
	if post.UserID != userID && role != "admin" {
		h.logger.Warn("Unauthorized post update attempt", "post_id", req.PostID)
		return utility.SendErrorResponse(c, http.StatusForbidden, "you can only update your own posts", errors.New("insufficient permission"))
	}

	updatedPost, err := h.postService.UpdatePost(req.PostID, req)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to update post", err)
		return utility.SendErrorResponse(c, statusCode, "failed to update post", err)
	}

	h.logger.Info("Post updated successfully", "post_id", req.PostID)
	return utility.SendSuccessResponse(c, http.StatusOK, "Post updated successfully", helper.PostToResponse(updatedPost))
}

// DeletePost handles post deletion
func (h *postHandler) DeletePost(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to DeletePost")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	postID := c.Params("id")
	id, err := uuid.Parse(postID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid post id", err)
	}

	post, err := h.postService.GetPostByID(id)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusNotFound, "failed to get post", err)
	}

	if post.UserID != userID {
		h.logger.Warn("Unauthorized post deletion attempt", "post_id", id)
		return utility.SendErrorResponse(c, http.StatusForbidden, "you can only delete your own posts", errors.New("insufficient permissions"))
	}

	err = h.postService.DeletePost(id)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to delete post", err)
		return utility.SendErrorResponse(c, statusCode, "failed to delete post", err)
	}

	h.logger.Info("Post deleted successfully", "post_id", id)
	return utility.SendSuccessResponse(c, http.StatusOK, "Post deleted successfully", nil)
}

// GetAdminStats handles getting admin statistics
func (h *postHandler) GetAdminStats(c fiber.Ctx) error {
	totalPosts, err1 := h.postService.GetTotalPostsCount()
	totalComments, err2 := h.commentService.GetTotalCommentsCount()
	firstPostDate, err3 := h.postService.GetFirstPostDate()

	if err1 != nil || err2 != nil || err3 != nil {
		h.logger.Error("Failed to get admin stats", err1)
		return utility.SendErrorResponse(c, http.StatusInternalServerError, "failed to get statistics", err1)
	}

	timeAgo := helper.CalculateTimeAgo(firstPostDate)

	stats := response.AdminOverviewResponse{
		TotalPosts:    totalPosts,
		TotalComments: totalComments,
		FirstPostAgo:  timeAgo,
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "Admin statistics retrieved successfully", stats)
}
