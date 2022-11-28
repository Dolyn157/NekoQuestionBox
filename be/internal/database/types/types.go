package types

import (
	"time"
)

type Question struct {
	ID         string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username   string     `json:"username"`
	Question   string     `gorm:"unique" json:"question"` // 每个问题只能出现一次
	Answer     *string    `json:"answer"`                 // 回答
	AnsweredAt *time.Time `json:"answeredAt"`             // 回答时间
	CreatedAt  time.Time  `json:"createdAt"`              // 创建时间
}

func (Question) TableName() string {
	return "public.questions"
}

type User struct {
	Username string `gorm:"primaryKey" json:"username"`
	Pwd      string `json:"pwd"`
}

func (User) TableName2() string {
	return "public.users"
}
