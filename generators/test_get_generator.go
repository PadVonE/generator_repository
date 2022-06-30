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


func GenerateTestGetCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct) (code string, err error) {
	
	path := filepath.FromSlash("./generators/template/test/_get_test.txt")
	pathFunc := filepath.FromSlash("./generators/template/test/_get/_func.txt")

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


	if len(pathFunc) > 0 && !os.IsPathSeparator(pathFunc[0]) {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		pathFunc = filepath.Join(wd, pathFunc)
	}

	datFunc, errFunc := ioutil.ReadFile(pathFunc)
	if errFunc != nil {
		log.Println(errFunc)
		return
	}
	sourceFunc := string(datFunc)

	tFunc := template.Must(template.New("const-list").Parse(sourceFunc))

	funcCode := ""

	name,_ := strc.NameInterface()

	listRequestElement,imports := generateGetRequestElement(strc.ResponseStruct)

	for _,rs := range strc.RequestStruct.Rows{

		data := DataTest{
			Name:           name,
			NameInSnake:    strcase.ToSnake(name),
			FilterBy:    	rs.Name,
			Imports:        imports,
			PackageStruct:  packageStruct,
			FinishedStruct: listRequestElement,
			TestList2:      generateEqualList("contents[1]","response",strc.ResponseStruct),
		}

		var funcTpl bytes.Buffer
		if err := tFunc.Execute(&funcTpl, data); err != nil {
			//return err
		}

		funcCode += "\n\n"+funcTpl.String()
	}


	data := DataTest{
		Name:           name,
		NameInSnake:    strcase.ToSnake(name),
		Imports:        imports,
		Functions:        funcCode,
		PackageStruct:  packageStruct,
		FinishedStruct: listRequestElement,
		TestList2:      generateEqualList("response","get",strc.ResponseStruct),
	}


	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()


	return
}

func generateGetRequestElement(p entity.Struct) (code string,imports string) {

	code = ""
	imports = ""

	for j := 0; j < 3; j++ {
		code += "\t\t{\n"
		for i, element := range p.Rows {
			if element.Name=="Id" ||
				element.Name=="CreatedAt" ||
				element.Name=="UpdatedAt" ||
				element.Name=="PublicDate" {
				continue
			}

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

