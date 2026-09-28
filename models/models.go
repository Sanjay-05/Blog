package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (base *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}

type User struct {
	Base
	UserName string `gorm:"type:text;not null" json:"user_name"`
	Email    string `gorm:"type:text;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`
	Role     string `gorm:"type:text;default:'user'" json:"role"`
	Status   string `gorm:"type:text;default:'active'" json:"status"`

	Posts    []Post    `gorm:"foreignKey:UserID" json:"posts,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}

type Category struct {
	Base
	CategoryName string    `gorm:"type:text;not null" json:"category_name"`
	Description  string    `gorm:"type:text" json:"description"`
	Status       string    `gorm:"type:text;default:'active'" json:"status"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`

	User  *User  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user,omitempty"`
	Posts []Post `gorm:"foreignKey:CategoryID" json:"posts,omitempty"`
}

type Post struct {
	Base
	Title   string `gorm:"type:text;not null" json:"title"`
	Slug    string `gorm:"type:text;not null" json:"slug"`
	Content string `gorm:"type:text;not null" json:"content"`
	Status  string `gorm:"type:text;default:'published'" json:"status"`

	UserID     uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`

	User     *User     `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user"`
	Category *Category `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"category"`

	Comments []Comment `gorm:"foreignKey:PostID" json:"comments,omitempty"`
}

type Comment struct {
	Base
	PostID  uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Comment string    `gorm:"type:text;not null"`

	ParentID *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	Replies  []Comment  `gorm:"foreignKey:ParentID" json:"replies,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"user"`
	Post Post `gorm:"foreignKey:PostID" json:"post"`
}

type JWTCustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}
