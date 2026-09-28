package repository

import (
	"blog/models"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines user data operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByName(name string) (*models.User, error)
	GetUserByEmail(name string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetAllUsers(limit, offset int) ([]models.User, int64, error)
	UpdateUser(id uuid.UUID, updates map[string]interface{}) error
	DeleteUser(id uuid.UUID) error
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// CreateUser creates a new user
func (r *userRepo) CreateUser(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// GetUserByEmail gets user by email
func (r *userRepo) GetUserByName(name string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("user_name = ? AND status = ?", name, "active").First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail gets user by email
func (r *userRepo) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.
		Where("email = ? AND status = ?", email, "active").
		First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID gets user by ID
func (r *userRepo) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.
		Where("id = ? AND status = ?", id, "active").
		First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetAllUsers gets all users with pagination
func (r *userRepo) GetAllUsers(limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if err := r.db.Model(&models.User{}).Where("status = ?", "active").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Where("status = ?", "active").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUser updates user information
func (r *userRepo) UpdateUser(id uuid.UUID, updates map[string]interface{}) error {
	result := r.db.Model(&models.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return result.Error
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to update user details")
	}
	return nil
}

// DeleteUser deletes a user
func (r *userRepo) DeleteUser(id uuid.UUID) error {
	now := time.Now()

	result := r.db.Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "deleted",
			"deleted_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("failed to delete user")
	}
	return nil
}
