package repository

import (
	"blog/models"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommentRepository defines comment data operations
type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	GetCommentByID(id uuid.UUID) (*models.Comment, error)
	GetCommentsByPost(postID uuid.UUID, limit, offset int) ([]models.Comment, int64, error)
	GetCommentsByUser(userID uuid.UUID, limit, offset int) ([]models.Comment, int64, error)
	UpdateComment(id uuid.UUID, updates map[string]interface{}) error
	DeleteComment(id uuid.UUID) error
	GetTotalCommentsCount() (int64, error)
	DeleteCommentsByPost(postID uuid.UUID) error
	GetReplies(parentID uuid.UUID, limit, offset int) ([]models.Comment, int64, error)
}

type commentRepo struct {
	db *gorm.DB
}

// NewCommentRepository creates a new comment repository
func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepo{db: db}
}

// CreateComment creates a new comment
func (r *commentRepo) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// GetCommentByID gets comment by ID
func (r *commentRepo) GetCommentByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	if err := r.db.Preload("User").Preload("Post").
		Where("id = ?", id).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("comment not found")
		}
		return nil, err
	}
	return &comment, nil
}

// GetCommentsByPost gets comments by post
func (r *commentRepo) GetCommentsByPost(postID uuid.UUID, limit, offset int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := r.db.Where("post_id = ? AND parent_id IS NULL", postID)

	if err := query.Model(&models.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").
		Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// GetCommentsByUser gets comments by user
func (r *commentRepo) GetCommentsByUser(userID uuid.UUID, limit, offset int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := r.db.Where("user_id = ?", userID)

	if err := query.Model(&models.Comment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").
		Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// UpdateComment updates comment information
func (r *commentRepo) UpdateComment(id uuid.UUID, updates map[string]interface{}) error {
	result := r.db.Model(&models.Comment{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return result.Error
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to update comment")
	}
	return nil
}

// DeleteComment deletes a comment
func (r *commentRepo) DeleteComment(id uuid.UUID) error {
	result := r.db.Where("id = ?", id).Delete(&models.Comment{})
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return result.Error
		}
		return result.Error
	}
	return nil
}

// GetTotalCommentsCount gets total comments count
func (r *commentRepo) GetTotalCommentsCount() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Comment{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteCommentsByPost deletes all comments for a post
func (r *commentRepo) DeleteCommentsByPost(postID uuid.UUID) error {
	result := r.db.Where("post_id = ?", postID).Delete(&models.Comment{})
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return result.Error
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to delete comment")
	}
	return nil
}

func (r *commentRepo) GetReplies(parentID uuid.UUID, limit, offset int) ([]models.Comment, int64, error) {
	var replies []models.Comment
	var total int64

	query := r.db.Model(&models.Comment{}).
		Where("parent_id = ?", parentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("User").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&replies).Error; err != nil {
		return nil, 0, err
	}

	return replies, total, nil
}