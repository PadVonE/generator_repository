package generators

import (
	"bytes"
	"generator/entity"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func GenerateGeneral(file string, packageStruct entity.PackageStruct, tableNames []string) (code string, err error) {

	path := filepath.FromSlash("./generators/repository/template/general/" + strings.ToLower(file) + ".txt")
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

	data := DataGeneralGenerator{
		PackageStruct: packageStruct,
		DropTableCode: DropTableCode(tableNames),
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func DropTableCode(tableNames []string) (code string) {
	code = ""

	for _, table := range tableNames {
		code += "\ts.Service.DB.Exec(\"DELETE FROM " + table + "\")\n"
	}
	return
}
