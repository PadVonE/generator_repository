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


func GenerateTestCreateCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct) (code string, err error) {
	
	path := filepath.FromSlash("./generators/template/test/_create_test.txt")
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

	listRequestElement,imports := generateCreateRequestElement(strc.RequestStruct)


	name,_ := strc.NameInterface()

	data := DataTest{
		Name:           name,
		NameInSnake:    strcase.ToSnake(name),
		Imports:        imports,
		PackageStruct:  packageStruct,
		FinishedStruct: listRequestElement,
		TestList1:      generateEqualList("create"+name+"Request","response",strc.RequestStruct),
		TestList2:      generateEqualList("response","get",strc.ResponseStruct),
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()


	return
}

func generateCreateRequestElement(p entity.Struct) (code string,imports string) {

	code = ""
	imports = ""

	for i, element := range	p.Rows{

		generatedCode,generatedImport := generateRowRequest(element.Name,element.Type,i)

		code+=generatedCode
		if !strings.Contains(imports, generatedImport){
			imports += generatedImport
		}

	}

	return
}

