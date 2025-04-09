package models

type User struct {
	FirstName string `json:"first_name" gorm:"type:varchar(100)"`
	LastName  string `json:"last_name" gorm:"type:varchar(100)"`
	Mail      string `json:"mail" gorm:"type:varchar(100);primaryKey;unique;not null"`
	// Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Maso               string `json:"-" gorm:"type:varchar(50);not null"`
	Role               string `json:"role" gorm:"type:varchar(50);not null"`
	ReadPosts          int    `json:"-" gorm:"type:int;default:0"`
	RunPosts           int    `json:"-" gorm:"type:int;default:0"`
	ReadRandomPosts    int    `json:"-" gorm:"type:int;default:0"`
	ReadSuggestedPosts int    `json:"-" gorm:"type:int;default:0"`
	RunSuggestedPosts  int    `json:"-" gorm:"type:int;default:0"`
	SearchPosts        int    `json:"-" gorm:"type:int;default:0"`
	ReadSearchPosts    int    `json:"-" gorm:"type:int;default:0"`
}
type GoogleUser struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}
