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

func GenerateTestUpdateCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct) (code string, err error) {

	path := filepath.FromSlash("./generator/generators/template/test/_update_test.txt")
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

	finishedStruct, imports := generateUpdateFinishedStruct(strc.RequestStruct)
	structForRequest, _ := generateUpdateRequestElements(strc.RequestStruct)

	name, _ := strc.NameInterface()

	data := DataTest{
		Name:             name,
		NameInSnake:      strcase.ToSnake(name),
		Imports:          imports,
		PackageStruct:    packageStruct,
		FinishedStruct:   finishedStruct,
		StructForRequest: structForRequest,
		TestList1:        generateEqualList("request", "response", strc.RequestStruct),
		TestList2:        generateEqualList("response", "get", strc.ResponseStruct),
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func generateUpdateFinishedStruct(p entity.Struct) (code string, imports string) {

	code = ""
	imports = ""

	for j := 0; j < 2; j++ {
		code += "\t\t{\n"
		for i, element := range p.Rows {
			generatedCode, generatedImport := generateRowRequest(element.Name, element.Type, i+(j*(len(p.Rows))))

			code += "\t" + generatedCode
			if !strings.Contains(imports, generatedImport) {
				imports += generatedImport
			}
		}
		code += "\t\t},\n"
	}

	return
}

func generateUpdateRequestElements(p entity.Struct) (code string, imports string) {

	code = ""
	imports = ""

	for i, element := range p.Rows {
		if element.Name=="Id" {
			continue
		}
		generatedCode, generatedImport := generateRowRequest(element.Name, element.Type, (i+1)*11)

		code += "\t" + generatedCode
		if !strings.Contains(imports, generatedImport) {
			imports += generatedImport
		}
	}

	return
}
