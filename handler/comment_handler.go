package handler

import (
	"blog/helper"
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

// CommentHandler handles comment-related requests
type CommentHandler interface {
	CreateComment(c fiber.Ctx) error
	GetComment(c fiber.Ctx) error
	GetCommentsByPost(c fiber.Ctx) error
	GetCommentsByUser(c fiber.Ctx) error
	UpdateComment(c fiber.Ctx) error
	DeleteComment(c fiber.Ctx) error
	AdminDeleteComment(c fiber.Ctx) error
}

type commentHandler struct {
	commentService service.CommentService
	logger         *logrus.Logger
}

// NewCommentHandler creates a new comment handler
func NewCommentHandler(cs service.CommentService, logger *logrus.Logger) CommentHandler {
	return &commentHandler{commentService: cs, logger: logger}
}

// CreateComment handles comment creation
func (h *commentHandler) CreateComment(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to CreateComment")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	var req request.CreateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to bind create comment request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(req); err != nil {
		h.logger.Error("Validation failed for create comment", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	comment, err := h.commentService.CreateComment(userID, req)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to create comment", err)
		return utility.SendErrorResponse(c, statusCode, "failed to create comment", err)
	}

	h.logger.Info("Comment Created successfully", "comment_id", comment.ID)
	return utility.SendSuccessResponse(c, http.StatusCreated, "comment created successfully", nil)
}

// GetComment handles getting a single comment
func (h *commentHandler) GetComment(c fiber.Ctx) error {
	commentID := c.Params("id")
	id, err := uuid.Parse(commentID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid comment id", err)
	}

	limit, offset := utility.GetPagination(c)
	comment, replies, total, err := h.commentService.GetCommentWithReplies(id, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get comment with replies", err)
		return utility.SendErrorResponse(c, http.StatusNotFound, "comment not found", err)
	}
	commentResponse := helper.CommentToResponse(comment)

	var replyResponses []response.CommentResponse
	for _, r := range replies {
		replyResponses = append(replyResponses, response.CommentResponse{
			ID:      r.ID,
			PostID:  r.PostID,
			UserID:  r.UserID,
			Comment: r.Comment,
			User: response.UserResponse{
				ID:       r.UserID,
				UserName: r.User.UserName,
			},
			CreatedAt: r.CreatedAt,
		})
	}
	return utility.SendSuccessResponse(c, http.StatusOK, "Comment retrieved successfully", map[string]interface{}{
		"comment": commentResponse,
		"replies": replyResponses,
		"pagination": map[string]interface{}{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetCommentsByPost handles getting comments by post
func (h *commentHandler) GetCommentsByPost(c fiber.Ctx) error {
	postID := c.Params("post_id")
	id, err := uuid.Parse(postID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid post id", err)
	}

	limit, offset := utility.GetPagination(c)

	comments, total, err := h.commentService.GetCommentsByPost(id, limit, offset)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to get comments", err)
		return utility.SendErrorResponse(c, statusCode, "failed to get comments", err)
	}

	if len(comments) == 0 {
		return utility.SendSuccessResponse(c, http.StatusOK, "No comments found", map[string]any{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"comments": []interface{}{},
		})
	}

	var commentResponses []response.CommentResponse
	for _, cmt := range comments {
		commentResponses = append(commentResponses, response.CommentResponse{
			ID:      cmt.ID,
			PostID:  cmt.PostID,
			UserID:  cmt.UserID,
			Comment: cmt.Comment,
			User: response.UserResponse{
				ID:       cmt.UserID,
				UserName: cmt.User.UserName,
			},
			CreatedAt: cmt.CreatedAt,
		})
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "Comments retrieved successfully", map[string]any{
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"comments": commentResponses,
	})
}

// GetCommentsByUser handles getting comments by user
func (h *commentHandler) GetCommentsByUser(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to get comments by user")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	limit, offset := utility.GetPagination(c)

	comments, total, err := h.commentService.GetCommentsByUser(userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get comments", err)
		return utility.SendErrorResponse(c, http.StatusInternalServerError, "failed to get comments", err)
	}

	if len(comments) == 0 {
		return utility.SendSuccessResponse(c, http.StatusOK, "No comments found", map[string]any{
			"total":    total,
			"limit":    limit,
			"offset":   offset,
			"comments": []interface{}{},
		})
	}

	var commentResponses []response.CommentResponse
	for _, cmt := range comments {
		commentResponses = append(commentResponses, response.CommentResponse{
			ID:      cmt.ID,
			PostID:  cmt.PostID,
			UserID:  cmt.UserID,
			Comment: cmt.Comment,
			User: response.UserResponse{
				ID:       cmt.UserID,
				UserName: cmt.User.UserName,
			},
			CreatedAt: cmt.CreatedAt,
		})
	}

	return utility.SendSuccessResponse(c, http.StatusOK, "comments retrieved successfully", map[string]any{
		"total":    total,
		"limit":    limit,
		"offset":   offset,
		"comments": commentResponses,
	})
}

// UpdateComment handles comment update
func (h *commentHandler) UpdateComment(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to UpdateComment")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	var req request.UpdateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		h.logger.Error("Failed to bind update comment request", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid request payload", err)
	}

	if err := helper.ValidateJson(&req); err != nil {
		h.logger.Error("Validation failed", err)
		return utility.SendErrorResponse(c, http.StatusBadRequest, "validation failed", err)
	}

	comment, err := h.commentService.UpdateComment(req.CommentID, userID, req)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to update comment", err)
		return utility.SendErrorResponse(c, statusCode, "failed to update comment", err)
	}

	h.logger.Info("Comment updated successfully", "comment_id", req.CommentID)
	return utility.SendSuccessResponse(c, http.StatusOK, "comment updated successfully", helper.CommentToResponse(comment))
}

// DeleteComment handles comment deletion (user can delete own comments)
func (h *commentHandler) DeleteComment(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		h.logger.Warn("Unauthorized access to DeleteComment")
		return utility.SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", errors.New("user not authenticated"))
	}

	commentID := c.Params("id")
	id, err := uuid.Parse(commentID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid comment id", err)
	}

	err = h.commentService.DeleteComment(id, userID)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to delete comment", err)
		return utility.SendErrorResponse(c, statusCode, "failed to delete comment", err)
	}

	h.logger.Info("comment deleted successfully", "comment_id", id)
	return utility.SendSuccessResponse(c, http.StatusOK, "Comment deleted successfully", nil)
}

// AdminDeleteComment handles comment deletion by admin
func (h *commentHandler) AdminDeleteComment(c fiber.Ctx) error {
	commentID := c.Params("id")
	id, err := uuid.Parse(commentID)
	if err != nil {
		return utility.SendErrorResponse(c, http.StatusBadRequest, "invalid comment id", err)
	}

	err = h.commentService.AdminDeleteComment(id)
	if err != nil {
		statusCode := utility.GetStatusCodeFromError(err)
		h.logger.Error("Failed to delete comment", err)
		return utility.SendErrorResponse(c, statusCode, "failed to delete comment", err)
	}

	h.logger.Info("comment deleted by admin", "comment_id", id)
	return utility.SendSuccessResponse(c, http.StatusOK, "comment deleted successfully", nil)
}
