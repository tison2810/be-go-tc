package models

type User struct {
	FirstName string `json:"first_name" gorm:"type:varchar(100);not null"`
	LastName  string `json:"last_name" gorm:"type:varchar(100);not null"`
	Mail      string `json:"mail" gorm:"type:varchar(100);primaryKey;unique;not null"`
	// Password  string    `json:"password" gorm:"type:varchar(255);not null"`
	Maso               string `json:"maso" gorm:"type:varchar(50);not null"`
	Role               string `json:"role" gorm:"type:varchar(50);not null"`
	ReadPosts          int    `json:"read_posts" gorm:"type:int;default:0"`
	RunPosts           int    `json:"run_posts" gorm:"type:int;default:0"`
	ReadRandomPosts    int    `json:"read_random_posts" gorm:"type:int;default:0"`
	ReadSuggestedPosts int    `json:"read_suggested_posts" gorm:"type:int;default:0"`
	RunSuggestedPosts  int    `json:"run_suggested_posts" gorm:"type:int;default:0"`
	SearchPosts        int    `json:"search_posts" gorm:"type:int;default:0"`
	ReadSearchPosts    int    `json:"read_search_posts" gorm:"type:int;default:0"`
}
type GoogleUser struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}
