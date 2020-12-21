package generators

import (
	"bytes"
	"generator/entity"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func GenerateServiceCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct,action string) (code string, err error) {

	path := filepath.FromSlash("./generator/generators/template/service/_"+strings.ToLower(action)+".txt")
	if len(path) > 0 && !os.IsPathSeparator(path[0]) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(wd, path)
	}

	dat, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return
	}
	source := string(dat)

	t := template.Must(template.New("const-list").Parse(source))


	name,_ := strc.NameInterface()

	data := Data{
		Name:          name,
		NameInSnake:   strcase.ToSnake(name),
		PackageStruct: packageStruct,

		ListFilter: ListFilter(strc.RequestStruct,strcase.ToSnake(name)),
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func ListFilter(request entity.Struct,nameInSnake string) (code string)  {
	code = ""
	for _, row := range request.Rows {

		switch row.Type {
			case "string":
				code += "\tif len(strings.TrimSpace(request."+row.Name+")) > 0 {\n"+
				"\t\tquery = query.Where(\"("+nameInSnake+".id::text = ? or lower("+nameInSnake+".name) like ?)\", strings.TrimSpace(request."+row.Name+"), \"%\"+strings.ToLower(strings.TrimSpace(request."+row.Name+"))+\"%\")\n"+
				"\t}\n"

			default:
				log.Warn("Type: " + row.Type + " not implemented (Generate Entity ListFilter)")
		}


	}



	return
}