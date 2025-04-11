package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	UserMail       string       `json:"mail" gorm:"type:varchar(100);not null"`
	Subject        string       `json:"subject" gorm:"type:varchar(255);not null"`
	Title          string       `json:"title" gorm:"type:varchar(255);not null"`
	Description    string       `json:"description" gorm:"type:text;not null"`
	LastModified   time.Time    `json:"last_modified" gorm:"autoCreateTime"`
	Trace          string       `json:"-" gorm:"type:varchar(255)"`
	IsDeleted      bool         `json:"-" gorm:"type:boolean;default:false"`
	Views          int          `json:"-" gorm:"type:int;default:0"`
	ViewsBySuggest int          `json:"-" gorm:"type:int;default:0"`
	ViewsBySearch  int          `json:"-" gorm:"type:int;default:0"`
	Runs           int          `json:"-" gorm:"type:int;default:0"`
	RunsBySuggest  int          `json:"-" gorm:"type:int;default:0"`
	Testcase       *Testcase    `json:"testcase" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Tags           []PostHasTag `json:"tags" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Testcase struct {
	PostID   uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	Input    string    `json:"input" gorm:"type:text;not null"`
	Expected string    `json:"expected" gorm:"type:text;not null"`
	Code     string    `json:"code" gorm:"type:text;not null"`

	Post *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Interaction struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;not null"`
	UserMail  string    `json:"user_mail" gorm:"type:varchar(100);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	IsLike    bool      `json:"is_like" gorm:"type:boolean;default:false"`

	Post *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User *User `json:"-" gorm:"foreignKey:UserMail;references:Mail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type StudentRunTestcase struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	PostID      uuid.UUID `json:"post_id" gorm:"type:uuid"`
	StudentMail string    `json:"student_mail" gorm:"type:varchar(100);primaryKey"`
	Log         string    `json:"log" gorm:"type:varchar(255)"`
	Score       int       `json:"score" gorm:"type:int"`
	Time        time.Time `json:"time" gorm:"autoCreateTime"`

	Post    *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Student *User `json:"-" gorm:"foreignKey:StudentMail;references:Mail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type TeacherVerifyPost struct {
	PostID      uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	TeacherMail string    `json:"teacher_mail" gorm:"type:varchar(100);not null"`

	Post    *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Teacher *User `json:"-" gorm:"foreignKey:TeacherMail;references:Mail;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}
type Tag struct {
	ID   int    `json:"id" gorm:"type:integer;primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"type:varchar(255);not null"`

	Posts []PostHasTag `json:"-" gorm:"foreignKey:TagID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type PostHasTag struct {
	PostID uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	TagID  int       `json:"tag_id" gorm:"type:integer;primaryKey"`

	Post *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Tag  *Tag  `json:"-" gorm:"foreignKey:TagID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type PostInteraction struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserMail  string    `json:"user_mail" gorm:"type:varchar(100);not null"`
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;not null"`
	PostType  int       `json:"post_type" gorm:"type:int;not null"`
	Action    string    `json:"action" gorm:"type:varchar(10);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	Post *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User *User `json:"-" gorm:"foreignKey:UserMail;references:Mail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type InteractionInfo struct {
	LikeCount           int64      `json:"like_count"`    // Số lượt like
	CommentCount        int64      `json:"comment_count"` // Số lượt comment
	LikeID              *uuid.UUID `json:"like_id"`       // ID của like nếu user đã like, null nếu chưa
	VerifiedTeacherMail *string    `json:"verified_teacher_mail"`
	Views               int        `json:"view_count"` // Số lượt xem
	Runs                int        `json:"run_count"`  // Số lượt chạy
}

type PostStats struct {
	PostID              uuid.UUID  `gorm:"column:post_id"`
	LikeCount           int64      `gorm:"column:like_count"`
	CommentCount        int64      `gorm:"column:comment_count"`
	LikeID              *uuid.UUID `gorm:"column:like_id"`
	VerifiedTeacherMail *string    `gorm:"column:verified_teacher_mail"`
	Views               int        `gorm:"column:views"`
	Runs                int        `gorm:"column:runs"`
}
