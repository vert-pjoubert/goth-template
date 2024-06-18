package models

type Server struct {
	ID    int64 `xorm:"pk autoincr"`
	Name  string
	Type  string
	URL   string
	Roles []string `xorm:"json"`
}

func (s *Server) TableName() string {
	return "servers"
}
