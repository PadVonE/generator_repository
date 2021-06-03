package usecase

import (
	"html/template"
)

func GetTemplate() (tmpl map[string]*template.Template) {
	tmpl = map[string]*template.Template{}

	listTemplates := []string{
		"index",
	}

	for _, tName := range listTemplates {
		tmpl[tName] = template.Must(template.ParseFiles(
			"./view/layout.html",
			"./view/_menu.html",
			"./view/"+tName+".html",
		))

	}

	return tmpl
}
