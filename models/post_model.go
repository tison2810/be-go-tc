package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID                  uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	UserMail            string         `json:"user_mail" gorm:"type:varchar(100);not null"`
	VerifiedTeacherMail sql.NullString `json:"verified_teacher_mail,omitempty" gorm:"type:varchar(100)"`
	Subject             string         `json:"subject" gorm:"type:varchar(255);not null"`
	Title               string         `json:"title" gorm:"type:varchar(255);not null"`
	Description         string         `json:"description" gorm:"type:text;not null"`
	CreatedAt           time.Time      `json:"created_at" gorm:"autoCreateTime"`
	Trace               string         `json:"trace" gorm:"type:varchar(255)"`

	Topics   []PostTopic `json:"topics" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Testcase *Testcase   `json:"testcase" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Testcase struct {
	PostID   uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	Input    string    `json:"input" gorm:"type:varchar(100);not null"`
	Expected string    `json:"expected" gorm:"type:varchar(100);not null"`

	Post *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type Interaction struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`              // Sử dụng UUID thay vì int để đồng nhất với các model khác
	PostID    uuid.UUID `json:"post_id" gorm:"type:uuid;not null"`           // Liên kết với Post.ID
	UserMail  string    `json:"user_mail" gorm:"type:varchar(100);not null"` // Liên kết với User.mail
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Type      string    `json:"type" gorm:"type:varchar(50);not null"`     // "Like" hoặc "Rating"
	Rating    int       `json:"rating" gorm:"type:int;default:0"`          // Điểm đánh giá, mặc định là 0
	IsLike    bool      `json:"is_like" gorm:"type:boolean;default:false"` // Trạng thái Like, mặc định là false

	Post *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	User *User `json:"-" gorm:"foreignKey:UserMail;references:Mail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type PostTopic struct {
	Topic  string    `json:"topic"`
	PostID uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	Post   *Post     `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
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
