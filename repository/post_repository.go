package repository

import (
	"blog/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostRepository defines post data operations
type PostRepository interface {
	CreatePost(post *models.Post) error
	GetPostByID(id uuid.UUID) (*models.Post, error)
	GetPostBySlug(slug string) (*models.Post, error)
	GetAllPosts(limit, offset int, status string) ([]models.Post, int64, error)
	GetPostsByCategory(categoryID uuid.UUID, limit, offset int) ([]models.Post, int64, error)
	GetPostsByUser(userID uuid.UUID, limit, offset int) ([]models.Post, int64, error)
	UpdatePost(id uuid.UUID, updates map[string]interface{}) error
	DeletePost(id uuid.UUID) error
	GetTotalPostsCount() (int64, error)
	GetFirstPostDate() (time.Time, error)
}

type postRepo struct {
	db *gorm.DB
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepo{db: db}
}

// CreatePost creates a new post
func (r *postRepo) CreatePost(post *models.Post) error {
	return r.db.Create(post).Error
}

// GetPostByID gets post by ID with relations
func (r *postRepo) GetPostByID(id uuid.UUID) (*models.Post, error) {
	var post models.Post
	if err := r.db.Preload("User").Preload("Category").Preload("Comments").
		Where("id = ?", id).First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

// GetPostBySlug finds an active post by slug
func (r *postRepo) GetPostBySlug(slug string) (*models.Post, error) {
	var post models.Post
	if err := r.db.Where("slug = ? AND status != ?", slug, "deleted").First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("post not found")
		}
		return nil, err
	}
	return &post, nil
}

// GetAllPosts gets all posts with pagination
func (r *postRepo) GetAllPosts(limit, offset int, status string) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.Model(&models.Post{})

	if status == "" {
		query = query.Where("status != ?", "deleted")
	} else {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").Preload("Category").
		Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsByCategory gets posts by category (exclude deleted)
func (r *postRepo) GetPostsByCategory(categoryID uuid.UUID, limit, offset int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.Where("category_id = ? AND status != ?", categoryID, "deleted")

	if err := query.Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").Preload("Category").
		Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// GetPostsByUser gets posts by user (exclude deleted)
func (r *postRepo) GetPostsByUser(userID uuid.UUID, limit, offset int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := r.db.Where("user_id = ? AND status != ?", userID, "deleted")

	if err := query.Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").Preload("Category").
		Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// UpdatePost updates post information
func (r *postRepo) UpdatePost(id uuid.UUID, updates map[string]interface{}) error {
	result := r.db.Model(&models.Post{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return result.Error
		}
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("failed to update post details")
	}
	return nil
}

// DeletePost deletes a post
func (r *postRepo) DeletePost(id uuid.UUID) error {
	result := r.db.Where("id = ?", id).Delete(&models.Post{})
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return result.Error
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to delete post details")
	}
	return nil
}

// GetTotalPostsCount gets total posts count
func (r *postRepo) GetTotalPostsCount() (int64, error) {
    var count int64
    if err := r.db.Model(&models.Post{}).Where("status != ?", "deleted").Count(&count).Error; err != nil {
        return 0, err
    }
    return count, nil
}

// GetFirstPostDate returns the created date of the first non-deleted post
func (r *postRepo) GetFirstPostDate() (time.Time, error) {
	var firstPost models.Post
	err := r.db.Where("status != ?", "deleted").
		Order("created_at asc").
		First(&firstPost).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	return firstPost.CreatedAt, nil
}