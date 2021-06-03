package entity

import "time"

type Project struct {
	Id               int32     `xorm:"'id' pk autoincr"`
	CreatedAt        time.Time `xorm:"created_at created"`
	UpdatedAt        time.Time `xorm:"updated_at updated"`
	GithubProto      string    `xorm:"github_proto"`
	FolderRepository string    `xorm:"folder_repository"`
	FolderName       string    `xorm:"folder_name"`
	Type             int32     `xorm:"type"`
	Version          string    `xorm:"version"`
	PackageInfo      string    `xorm:"package_info"`
	ListOfStruct     string    `xorm:"list_of_struct"`
	FuncList         string    `xorm:"func_list"`
}

func (project *Project) TableName() string {
	return "project"
}
