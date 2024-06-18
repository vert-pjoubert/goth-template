package models

import "time"

type User struct {
	ID        int64  `xorm:"pk autoincr"`
	Email     string `xorm:"unique"`
	Name      string
	RoleID    int64     `xorm:"index"`
	Role      Role      `xorm:"-"`
	CreatedAt time.Time `xorm:"created"`
	UpdatedAt time.Time `xorm:"updated"`
}

func (u *User) TableName() string {
	return "users"
}