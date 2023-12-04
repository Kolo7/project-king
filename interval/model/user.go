package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint         `gorm:"primarykey"`
	Username    string       `gorm:"uniqueIndex;type:varchar(255)" json:"username"`
	Password    string       `gorm:"not null;type:varchar(255)" json:"-"` // 这里假设需要保护用户密码，因此在 JSON 输出中不包括 Password 字段
	Phone       string       `gorm:"uniqueIndex;type:varchar(255)" json:"phone"`
	Slat        string       `gorm:"not null;type:varchar(255)" json:"-"`
	Email       string       `gorm:"not null;type:varchar(255)" json:"email"`
	Nickname    string       `gorm:"not null;type:varchar(255)" json:"nickname"`
	Sex         string       `gorm:"not null;type:varchar(255)" json:"sex"`
	Province    string       `gorm:"not null" json:"province"`
	City        string       `gorm:"not null" json:"city"`
	Remark      string       `gorm:"not null" json:"remark"`
	Avatarurl   string       `gorm:"not null" json:"avatarurl"`
	Industry    string       `gorm:"not null" json:"industry"`
	Company     string       `gorm:"not null" json:"company"`
	Workage     string       `gorm:"not null" json:"workage"`
	Title       string       `gorm:"not null" json:"title"`
	Status      string       `gorm:"not null" json:"status"`
	Token       int64        `gorm:"not null" json:"token"`
	CreatedAt   time.Time    `gorm:"column:createTime"`
	UpdatedAt   time.Time    `gorm:"column:updatedTime"`
	UserSecrets []UserSecret `gorm:"foreignKey:UserId"`
}

type UserSecret struct {
	gorm.Model
	ApiKey    string `gorm:"column:api_key;uniqueIndex;type:varchar(255)"`
	UserId    uint   `gorm:"column:user_id"`
	ApiSecret string `gorm:"column:api_secret;type:varchar(255)"`
	Charge    int64  `gorm:"column:charge"`
	User      User   `gorm:"foreignKey:UserId"`
}

// 反序列化user
func (u *User) UnmarshalBinary(data string) error {
	return json.Unmarshal([]byte(data), u)
}

// 序列化user
func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}
