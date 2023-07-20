package generators

import (
	"bytes"
	"generator/entity"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func GenerateTestCreateCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct, nameInterface entity.NameInterface) (code string, err error) {

	path := filepath.FromSlash("./generators/repository/template/test/_create_test.txt")
	if len(path) > 0 && !os.IsPathSeparator(path[0]) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(wd, path)
	}

	dat, err := os.ReadFile(path)
	if err != nil {
		log.Error(err)
		return
	}
	source := string(dat)

	t := template.Must(template.New("const-list").Parse(source))

	listRequestElement, imports := generateCreateRequestElement(strc.RequestStruct)

	data := DataTest{
		Name:           nameInterface.GetMethodName(),
		NameInSnake:    strcase.ToSnake(nameInterface.Name),
		Imports:        imports,
		PackageStruct:  packageStruct,
		FinishedStruct: listRequestElement,
		TestList1:      generateEqualList("create"+nameInterface.Name+"Request", "response", strc.RequestStruct),
		TestList2:      generateEqualList("response", "protoGet", strc.ResponseStruct),
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func generateCreateRequestElement(p entity.Struct) (code string, imports string) {

	code = ""
	imports = ""

	for i, element := range p.Rows {

		generatedCode, generatedImport := generateTestRowRequest(element.Name, element.Type, i, true)

		code += generatedCode
		if !strings.Contains(imports, generatedImport) {
			imports += generatedImport
		}

	}

	return
}
