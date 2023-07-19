package generators

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"html/template"
	"os"
	"path/filepath"
)

type Data struct {
	ServiceName string
	ServicePort string
}

func GenerateDockerComposeDatabase(projectName string, projectPort string) (code string, err error) {

	path := filepath.FromSlash("./generators/docker/template/gateway/_database.txt")

	if len(path) > 0 && !os.IsPathSeparator(path[0]) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(wd, path)
	}

	dat, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return
	}
	source := string(dat)

	t := template.Must(template.New("const-list").Parse(source))

	data := Data{
		ServiceName: projectName,
		ServicePort: projectPort,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}
