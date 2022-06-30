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
	
	path := filepath.FromSlash("./generators/template/test/_list_test.txt")
	pathFunc := filepath.FromSlash("./generators/template/test/_list/_func.txt")

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
		realisation := "// TODO implement test"

		switch rs.Type {
		case "int32":
			switch rs.Name {
			case "Limit":
				realisation = testByLimit(packageStruct,strc.BasicStruct,name)
			case "Offset":
				realisation = testByOffset(packageStruct,strc.BasicStruct,name)
			default:
				realisation = testByOtherInt(packageStruct,strc.BasicStruct,name,rs.Name)
			}
		case "string":
			switch rs.Name {
			case "Search":
				realisation = testByLimit(packageStruct,strc.BasicStruct,name)
			default:
				log.Println("List generator: not implemented test on field and type",rs.Name,rs.Type)
				realisation = testTemplate(packageStruct,strc.BasicStruct,name,rs.Name)
			}
		default:
			realisation = testTemplate(packageStruct,strc.BasicStruct,name,rs.Name)
		}

		data := DataTest{
			Name:           name,
			NameInSnake:    strcase.ToSnake(name),
			FilterBy:    	rs.Name,
			Imports:        imports,
			PackageStruct:  packageStruct,
			FinishedStruct: listRequestElement,
			TestList2:      generateEqualList("contents[1]","response",strc.BasicStruct),
			RealisationTest: realisation,
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
}

func testByLimit(packageStruct entity.PackageStruct,p entity.Struct,name string) (code string) {

	code += "\tlist"+name+"Request := &"+packageStruct.PackageName+".List"+name+"Request{\n"
	code += "\t\tLimit: 2,\n"
	code += "\t}\n"

	code += "\tresponse, err := s.Service.List"+name+"(context.Background(), list"+name+"Request)\n"
	code += "\tif !s.NoError(err) {	return }\n"
	code += "\tif !s.Equal(int32(3), response.Total) {	return }\n"
	code += "\tif !s.Equal(2, len(response.Items)) {	return }\n"
	code += "\n\n"


	code += generateEqualList("contents[2]","response.Items[0]",p)
	code += "\n\n"

	code += generateEqualList("contents[1]","response.Items[1]",p)

	return

}

func testByOffset(packageStruct entity.PackageStruct,p entity.Struct,name string) (code string) {

	code += "\tlist"+name+"Request := &"+packageStruct.PackageName+".List"+name+"Request{\n"
	code += "\t\tLimit: 2,\n"
	code += "\t\tOffset: 1,\n"
	code += "\t}\n"

	code += "\tresponse, err := s.Service.List"+name+"(context.Background(), list"+name+"Request)\n"
	code += "\tif !s.NoError(err) {	return }\n"
	code += "\tif !s.Equal(int32(3), response.Total) {	return }\n"
	code += "\tif !s.Equal(2, len(response.Items)) {	return }\n"
	code += "\n\n"

	code += generateEqualList("contents[1]","response.Items[0]",p)
	code += "\n\n"
	code += generateEqualList("contents[0]","response.Items[1]",p)

	return

}

func testByOtherInt(packageStruct entity.PackageStruct,p entity.Struct,name string,rowName string) (code string) {

	code += "\tlist"+name+"Request := &"+packageStruct.PackageName+".List"+name+"Request{\n"
	code += "\t\t"+rowName+": contents[1]."+rowName+",\n"
	code += "\t}\n"

	code += "\tresponse, err := s.Service.List"+name+"(context.Background(), list"+name+"Request)\n"
	code += "\tif !s.NoError(err) {	return }\n"
	code += "\tif !s.Equal(1, len(response.Items)) {	return }\n"
	code += "\n\n"

	code += generateEqualList("contents[1]","response.Items[0]",p)

	return
}

func testTemplate(packageStruct entity.PackageStruct,p entity.Struct,name string,rowName string) (code string) {
	code += "/* \n"

	code += "\tlist"+name+"Request := &"+packageStruct.PackageName+".List"+name+"Request{\n"
	code += "\t\t// TODO implement conditions\n"
	code += "\t}\n"

	code += "\tresponse, err := s.Service.List"+name+"(context.Background(), list"+name+"Request)\n"
	code += "\tif !s.NoError(err) {	return }\n"
	code += "\tif !s.Equal(1, len(response.Items)) {	return }\n"
	code += "\n\n"

	code += generateEqualList("contents[]","response.Items[]",p)
	code += "*/ \n"
	return
}



func testBySearch(packageStruct entity.PackageStruct,p entity.Struct,name string) (code string) {

	code += "\tlist"+name+"Request := &"+packageStruct.PackageName+".List"+name+"Request{\n"
	code += "\t\tLimit: 2,\n"
	code += "\t\tSearch: contents[1].Title[10:30],\n"
	code += "\t}\n"

	code += "\tresponse, err := s.Service.List"+name+"(context.Background(), list"+name+"Request)\n"
	code += "\tif !s.NoError(err) {	return }\n"
	code += "\tif !s.Equal(int32(1), response.Total) {	return }\n"
	code += "\tif !s.Equal(1, len(response.Items)) {	return }\n"
	code += "\n\n"

	code += generateEqualList("contents[1]","response.Items[0]",p)
	code += "\n\n"


	code += "\tlist"+name+"Request := &"+packageStruct.PackageName+".List"+name+"Request{\n"
	code += "\t\tLimit: 2,\n"
	code += "\t\tSearch: fmt.Sprintf(\"%d\", contents[2].Id),\n"
	code += "\t}\n"

	code += "\tresponse, err := s.Service.List"+name+"(context.Background(), list"+name+"Request)\n"
	code += "\tif !s.NoError(err) {	return }\n"
	code += "\tif !s.Equal(int32(1), response.Total) {	return }\n"
	code += "\tif !s.Equal(1, len(response.Items)) {	return }\n"
	code += "\n\n"
	code += generateEqualList("contents[2]","response.Items[0]",p)

	return

}


