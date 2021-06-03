package service

import (
	"errors"
	"github.com/go-xorm/xorm"
	"html/template"
)

var ErrNotFound = errors.New("not found")

type Service struct {
	DB *xorm.Engine
	Templates   map[string]*template.Template
}


func New() *Service {
	return &Service{}
}
