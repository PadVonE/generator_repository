package entity

import "time"

type AccessData struct {
	Id             int32     `xorm:"'id' pk autoincr"`
	CreatedAt      time.Time `xorm:"created_at created"`
	UpdatedAt      time.Time `xorm:"updated_at updated"`
	GithubUsername string    `xorm:"github_username"`
	GithubToken    string    `xorm:"github_token"`
}

func (accessData *AccessData) TableName() string {
	return "access_data"
}
