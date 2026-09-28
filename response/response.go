package response

import (
	"github.com/google/uuid"
	"time"
)

// Generic Response
type APIResponse struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	StatusCode int         `json:"status_code"`
	Error      string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Data   interface{} `json:"data"`
}

// User Response
type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	UserName string    `json:"user_name"`
	Email    string    `json:"email,omitempty"`
	Role     string    `json:"role,omitempty"`
}

type LoginResponse struct {
	Token uuid.UUID    `json:"token"`
	User  UserResponse `json:"user"`
}

// Category Response
type CategoryResponse struct {
	ID           uuid.UUID `json:"id"`
	CategoryName string    `json:"category_name"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

// Post Response
type PostResponse struct {
	ID         uuid.UUID         `json:"id"`
	Title      string            `json:"title"`
	Slug       string            `json:"slug"`
	Content    string            `json:"content"`
	Status     string            `json:"status"`
	UserID     *uuid.UUID        `json:"user_id,omitempty"`
	CategoryID *uuid.UUID        `json:"category_id,omitempty"`
	User       *UserResponse     `json:"user,omitempty"`
	Category   *CategoryResponse `json:"category,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

type PostDetailResponse struct {
	Post     PostResponse      `json:"post"`
	Comments []CommentResponse `json:"comments"`
}

// Comment Response
type CommentResponse struct {
	ID        uuid.UUID    `json:"id"`
	PostID    uuid.UUID    `json:"post_id"`
	UserID    uuid.UUID    `json:"user_id"`
	Comment   string       `json:"comment"`
	User      UserResponse `json:"user"`
	CreatedAt time.Time    `json:"created_at"`
}

// Admin Overview Response
type AdminOverviewResponse struct {
	TotalPosts      int64  `json:"total_posts"`
	TotalComments   int64  `json:"total_comments"`
	FirstPostAgo    string `json:"first_post"`
	TotalUsers      int64  `json:"total_users,omitempty"`
	TotalCategories int64  `json:"total_categories,omitempty"`
}

type CategoryWithCountResponse struct {
	ID           uuid.UUID `json:"id"`
	CategoryName string    `json:"category_name"`
	Description  string    `json:"description"`
	Count        int64     `json:"count"`
	CreatedAt    time.Time `json:"created_at"`
}
