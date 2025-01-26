package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserMail  string    `json:"user_mail" gorm:"type:varchar(100);not null"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;not null"`
	Post      *Post     `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Content   string    `json:"content,omitempty" gorm:"type:text"`
	ParentID  uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
