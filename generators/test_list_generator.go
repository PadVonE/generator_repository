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


func GenerateTestListCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct) (code string, err error) {
	
	path := filepath.FromSlash("./generator/generators/template/test/_list_test.txt")
	pathFunc := filepath.FromSlash("./generator/generators/template/test/_list/_func.txt")

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

	listRequestElement,imports := generateGetRequestElement(strc.BasicStruct)

	for _,rs := range strc.RequestStruct.Rows{

		data := DataTest{
			Name:           name,
			NameInSnake:    strcase.ToSnake(name),
			FilterBy:    	rs.Name,
			Imports:        imports,
			PackageStruct:  packageStruct,
			FinishedStruct: listRequestElement,
			TestList2:      generateEqualList("contents[1]","response",strc.BasicStruct),
		}

		var funcTpl bytes.Buffer
		if err := tFunc.Execute(&funcTpl, data); err != nil {
			//return err
		}

		funcCode += funcTpl.String()
	}


	data := DataTest{
		Name:           name,
		NameInSnake:    strcase.ToSnake(name),
		Imports:        imports,
		Functions:        funcCode,
		PackageStruct:  packageStruct,
		FinishedStruct: listRequestElement,
		TestList2:      generateEqualList("response","get",strc.BasicStruct),
	}


	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()


	return
}

func generateListRequestElement(p entity.Struct) (code string,imports string) {

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
	return
}

