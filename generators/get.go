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

func GenerateServiceCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct, action string) (code string, err error) {

	path := filepath.FromSlash("./generators/template/service/_" + strings.ToLower(action) + ".txt")
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

	name, _ := strc.NameInterface()


	listFilter,imports := ListFilter(strc.RequestStruct, strcase.ToSnake(name))
	data := Data{
		Name:          name,
		NameInSnake:   strcase.ToSnake(name),
		NameInCamel:   strcase.ToLowerCamel(name),
		PackageStruct: packageStruct,

		ListFilter: listFilter,
		Imports: imports,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func ListFilter(request entity.Struct, nameInSnake string) (code string, imports string) {
	code = ""
	imports = ""
	for _, row := range request.Rows {

		//Исключаем из фильтрации лимит и офсет
		if row.Name == "Limit" || row.Name == "Offset" {
			continue
		}

		switch row.Type {
		case "string":
			code += "\n\t// TODO Проверить правильно ли работает поиск " + row.Name + "(" + strcase.ToSnake(row.Name) + ") in table " + nameInSnake + "\n"
			code += "\tif len(strings.TrimSpace(request." + row.Name + ")) > 0 {\n" +
				"\t\tquery = query.Where(\"(" + nameInSnake + ".id::text = ? or lower(" + nameInSnake + ".name) like ?)\", strings.TrimSpace(request." + row.Name + "), \"%\"+strings.ToLower(strings.TrimSpace(request." + row.Name + "))+\"%\")\n" +
				"\t}\n\n"
			imports += "\t\"strings\""
		case "int32", "int64":
			code += "\tif request." + row.Name + " != 0 {\n" +
				"\t\tquery = query.Where(\"" + nameInSnake + "." + strcase.ToSnake(row.Name) + "\", request." + row.Name + ")\n" +
				"\t}\n\n"

		default:
			code += "\n\n"
			code += "\t// TODO not implemented " + row.Name + "(" + strcase.ToSnake(row.Name) + ") in table " + nameInSnake
			code += "\n\n"
			log.Warn("Type: " + row.Type + " not implemented (Generate Entity ListFilter)")
		}
	}

	return
}
