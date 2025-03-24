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
	Testcase            *Testcase      `json:"testcase" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Topics              []PostTopic    `json:"topics" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Trace               string         `json:"trace" gorm:"type:varchar(255)"`
}

type Testcase struct {
	PostID   uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	Post     *Post     `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Input    string    `json:"input" gorm:"type:varchar(100);not null"`
	Expected string    `json:"expected" gorm:"type:varchar(100);not null"`
}

type PostTopic struct {
	Topic  string    `json:"topic"`
	PostID uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	Post   *Post     `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// type File struct {
// 	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
// 	StudentMail string    `json:"student_mail" gorm:"type:varchar(100);not null"`
// 	Title       string    `json:"title" gorm:"type:varchar(100)"`
// 	Path        string    `json:"path" gorm:"type:varchar(255)"`
// 	Student     *User     `json:"-" gorm:"foreignKey:StudentMail;references:Mail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
// }

type StudentRunTestcase struct {
	PostID      uuid.UUID `json:"post_id" gorm:"type:uuid;primaryKey"`
	FileID      uuid.UUID `json:"file_id" gorm:"type:uuid;primaryKey"`
	StudentMail string    `json:"student_mail" gorm:"type:varchar(100);primaryKey"`
	Log         string    `json:"log" gorm:"type:varchar(255)"`
	Score       int       `json:"score" gorm:"type:int"`
	Time        time.Time `json:"time" gorm:"autoCreateTime"`

	Post *Post `json:"-" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	// File   *File `json:"-" gorm:"foreignKey:FileID,StudentMail;references:ID,StudentMail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Student *User `json:"-" gorm:"foreignKey:StudentMail;references:Mail;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
