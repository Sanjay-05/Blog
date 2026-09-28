package service

import (
	"blog/models"
	"blog/repository"
	"blog/request"
	"errors"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// CommentService defines comment business logic
type CommentService interface {
	CreateComment(userID uuid.UUID, req request.CreateCommentRequest) (*models.Comment, error)
	GetCommentByID(id uuid.UUID) (*models.Comment, error)
	GetCommentsByPost(postID uuid.UUID, limit, offset int) ([]models.Comment, int64, error)
	GetCommentsByUser(userID uuid.UUID, limit, offset int) ([]models.Comment, int64, error)
	UpdateComment(id uuid.UUID, userID uuid.UUID, req request.UpdateCommentRequest) (*models.Comment, error)
	DeleteComment(id uuid.UUID, userID uuid.UUID) error
	GetTotalCommentsCount() (int64, error)
	AdminDeleteComment(id uuid.UUID) error
	GetCommentWithReplies(id uuid.UUID, limit, offset int) (*models.Comment, []models.Comment, int64, error)
}

type commentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
	logger      *logrus.Logger
}

// NewCommentService creates a new comment service
func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
	logger *logrus.Logger,
) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// CreateComment creates a new comment
func (s *commentService) CreateComment(userID uuid.UUID, req request.CreateCommentRequest) (*models.Comment, error) {
	_, err := s.postRepo.GetPostByID(req.PostID)
	if err != nil {
		s.logger.Warn("Failed to create comment - post not found", "post_id", req.PostID)
		return nil, ErrForeignKeyViolation
	}

	_, err = s.userRepo.GetUserByID(userID)
	if err != nil {
		s.logger.Warn("Failed to create comment - user not found", "user_id", userID)
		return nil, ErrForeignKeyViolation
	}

	if req.ParentID != nil {
		parent, err := s.commentRepo.GetCommentByID(*req.ParentID)
		if err != nil {
			s.logger.Warn("Parent comment not found", "parent_id", *req.ParentID)
			return nil, ErrNotFound
		}
		if parent.PostID != req.PostID {
			s.logger.Warn("Reply post mismatch",
				"parent_post_id", parent.PostID,
				"req_post_id", req.PostID,
			)
			return nil, errors.New("reply must belong to same post")
		}
	}

	comment := &models.Comment{
		PostID:   req.PostID,
		UserID:   userID,
		Comment:  req.Comment,
		ParentID: req.ParentID,
	}

	if err := s.commentRepo.CreateComment(comment); err != nil {
		s.logger.Error("Failed to create comment", err)
		return nil, err
	}

	s.logger.Info("Comment created successfully", "comment_id", comment.ID)

	return s.commentRepo.GetCommentByID(comment.ID)
}

// GetCommentByID gets comment by ID
func (s *commentService) GetCommentByID(id uuid.UUID) (*models.Comment, error) {
	return s.commentRepo.GetCommentByID(id)
}

// GetCommentsByPost gets comments by post
func (s *commentService) GetCommentsByPost(postID uuid.UUID, limit, offset int) ([]models.Comment, int64, error) {
	_, err := s.postRepo.GetPostByID(postID)
	if err != nil {
		s.logger.Warn("Failed to get comment by post - post not found", "post_id", postID)
		return nil, 0, ErrForeignKeyViolation
	}
	return s.commentRepo.GetCommentsByPost(postID, limit, offset)
}

// GetCommentsByUser gets comments by user
func (s *commentService) GetCommentsByUser(userID uuid.UUID, limit, offset int) ([]models.Comment, int64, error) {
	_, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		s.logger.Warn("Failed to get comment by user - user not found", "user_id", userID)
		return nil, 0, ErrForeignKeyViolation
	}
	return s.commentRepo.GetCommentsByUser(userID, limit, offset)
}

// UpdateComment updates a comment
func (s *commentService) UpdateComment(id uuid.UUID, userID uuid.UUID, req request.UpdateCommentRequest) (*models.Comment, error) {
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		s.logger.Warn("Failed to update comment - comment not found", "comment_id", id)
		return nil, ErrNotFound
	}
	if req.Comment == "" {
		s.logger.Warn("Failed to update comment - empty comment field", "comment_id", id)
		return nil, errors.New("comment field cannot be empty")
	}
	if comment.UserID != userID {
		s.logger.Warn("Failed to update comment - unauthorized user", "comment_id", id, "user_id", userID)
		return nil, ErrOwnershipViolation
	}

	if err := s.commentRepo.UpdateComment(id, map[string]interface{}{
		"comment": req.Comment,
	}); err != nil {
		s.logger.Error("Failed to update comment", err)
		return nil, err
	}

	s.logger.Info("Comment updated successfully", "comment_id", id)
	return s.commentRepo.GetCommentByID(id)
}

// DeleteComment deletes a comment (only by owner)
func (s *commentService) DeleteComment(id uuid.UUID, userID uuid.UUID) error {
	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		s.logger.Warn("Failed to delete comment - comment not found", "comment_id", id)
		return ErrNotFound
	}

	if comment.UserID != userID {
		s.logger.Warn("Failed to delete comment - unauthorized user", "comment_id", id, "user_id", userID)
		return ErrOwnershipViolation
	}

	if err := s.commentRepo.DeleteComment(id); err != nil {
		s.logger.Error("Failed to delete comment", err)
		return err
	}

	s.logger.Info("Comment deleted successfully", "comment_id", id)
	return nil
}

// AdminDeleteComment deletes a comment (admin only)
func (s *commentService) AdminDeleteComment(id uuid.UUID) error {
	_, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		s.logger.Warn("Failed to delete comment - comment not found", "comment_id", id)
		return ErrNotFound
	}

	if err := s.commentRepo.DeleteComment(id); err != nil {
		s.logger.Error("Failed to delete comment", err)
		return err
	}

	s.logger.Info("Comment deleted by admin", "comment_id", id)
	return nil
}

// GetTotalCommentsCount gets total comments count
func (s *commentService) GetTotalCommentsCount() (int64, error) {
	return s.commentRepo.GetTotalCommentsCount()
}

func (s *commentService) GetCommentWithReplies(id uuid.UUID, limit, offset int) (*models.Comment, []models.Comment, int64, error) {

	comment, err := s.commentRepo.GetCommentByID(id)
	if err != nil {
		s.logger.Warn("comment not found", "comment_id", id)
		return nil, nil, 0, ErrNotFound
	}

	replies, total, err := s.commentRepo.GetReplies(id, limit, offset)
	if err != nil {
		s.logger.Error("failed to get replies", err)
		return nil, nil, 0, err
	}

	return comment, replies, total, nil
}