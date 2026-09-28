package request

import (
	"github.com/google/uuid"
)

// User Requests
type RegisterUserRequest struct {
	UserName string `json:"user_name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin|user"`
}

type LoginRequest struct {
	UserName string `json:"user_name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// Category Requests
type CreateCategoryRequest struct {
	CategoryName string `json:"category_name" validate:"required"`
	Description  string `json:"description" validate:"required"`
}

type UpdateCategoryRequest struct {
	CategoryID   uuid.UUID `json:"category_id" validate:"required,uuid4"`
	CategoryName *string   `json:"category_name"`
	Description  *string   `json:"description"`
}

// Post Requests
type CreatePostRequest struct {
	Title      string    `json:"title" validate:"required"`
	Slug       string    `json:"-"`
	Content    string    `json:"content" validate:"required"`
	CategoryID uuid.UUID `json:"category_id" validate:"required,uuid4"`
	Status string `json:"status" validate:"omitempty,oneof=published|draft"`
}

type UpdatePostRequest struct {
	PostID     uuid.UUID  `json:"post_id" validate:"required,uuid4"`
	Title      *string    `json:"title"`
	Slug       *string    `json:"slug"`
	Content    *string    `json:"content"`
	CategoryID *uuid.UUID `json:"category_id"`
	Status     *string    `json:"status" validate:"omitempty,oneof=published|draft"`
}

// Comment Requests
type CreateCommentRequest struct {
	PostID   uuid.UUID  `json:"post_id" validate:"required"`
	Comment  string     `json:"comment" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type UpdateCommentRequest struct {
	CommentID uuid.UUID `json:"comment_id" validate:"required,uuid4"`
	Comment   string    `json:"comment" validate:"required,min=1,max=500"`
}

type UpdateUserRequest struct {
	UserName string `json:"user_name" validate:"min=3,max=20"`
	Email    string `json:"email"`
}

type ForgetPasswordRequest struct {
	UserName string `json:"user_name" validate:"required"`
}

type ResetPasswordRequest struct {
	UserName           string `json:"user_name" validate:"required"`
	OTP                string `json:"otp" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=6"`
	ConfirmNewPassword string `json:"confirm_new_password" validate:"required,min=6"`
}

type SwitchRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=admin user"`
}