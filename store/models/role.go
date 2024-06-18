package models

type Role struct {
	ID          int64  `xorm:"pk autoincr"`
	Name        string `xorm:"unique"`
	Description string
	Permissions []string `xorm:"json"`
}

func (r *Role) TableName() string {
	return "roles"
}
