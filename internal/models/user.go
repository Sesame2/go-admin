package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username  string         `gorm:"type:varchar(255);not null;unique" json:"username"`
	Email     string         `gorm:"type:varchar(255);not null;unique" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string         `gorm:"type:varchar(255)" json:"nickname"`
	Role      string         `gorm:"type:varchar(50);not null;default:'user'" json:"role"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Documents []Document `gorm:"foreignKey:UserID" json:"documents,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
