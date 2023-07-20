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

func GenerateServiceCode(strc entity.ProtoInterfaceMethod, packageStruct entity.PackageStruct, nameInterface entity.NameInterface) (code string, err error) {

	suffix := ""
	for _, request := range strc.RequestStruct.Rows {
		if request.Name == "Uuid" {
			suffix = "_by_uuid"
		}
	}

	templateName := strings.ToLower(nameInterface.Action)

	if templateName == "update" || templateName == "delete" {
		templateName = templateName + suffix
	}

	path := filepath.FromSlash("./generators/repository/template/service/_" + templateName + ".txt")
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

	listFilter, imports := ListFilter(strc.RequestStruct, strcase.ToSnake(nameInterface.Name))
	data := Data{
		Name:          nameInterface.GetMethodName(),
		NameInSnake:   strcase.ToSnake(nameInterface.Name),
		NameInCamel:   strcase.ToLowerCamel(nameInterface.Name),
		PackageStruct: packageStruct,

		ListFilter: listFilter,
		Imports:    imports,
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

	countImports := map[string]int{}
	//usePrefixTable :=  nameInSnake + "."
	usePrefixTable := ""

	for _, row := range request.Rows {

		//Исключаем из фильтрации лимит и офсет
		if row.Name == "Limit" || row.Name == "Offset" {
			continue
		}

		switch row.Type {
		case "string":

			if countImports["string"] == 0 {
				imports += "\t\"strings\""
			}
			countImports["string"]++

			if row.Name == "Uuid" || strings.Contains(row.Name, "Uuid") {
				code += "\tif len(request." + row.Name + ") > 0 {\n" +
					"\t\tquery = query.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " = ?\", request." + row.Name + ")\n" +
					"\t}\n\n"
				break
			}

			code += "\n\t// TODO Проверить правильно ли работает поиск " + row.Name + "(" + strcase.ToSnake(row.Name) + ") in table " + nameInSnake + "\n"

			code += "\tif len(strings.TrimSpace(request." + row.Name + ")) > 0 {\n" +
				"\t\tquery = query.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " ilike ?\", \"%\"+request." + row.Name + "+\"%\")\n" +
				"\t}\n\n"
			if countImports["string"] == 0 {
				imports += "\t\"strings\""
			}
			countImports["string"]++
		case "int32", "int64", "float32", "float64":
			code += "\tif request." + row.Name + " > 0 {\n" +
				"\t\tquery = query.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " = ?\", request." + row.Name + ")\n" +
				"\t}\n\n"
		case "[]int32":
			trimmed := strings.TrimSuffix(strcase.ToSnake(row.Name), "s")
			code += "\tif len(request." + row.Name + ") > 0 {\n" +
				"\t\tquery = query.Where(\"" + usePrefixTable + "" + trimmed + " IN ?\", request." + row.Name + ")\n" +
				"\t}\n\n"
		case "[]string":
			trimmed := strings.TrimSuffix(strcase.ToSnake(row.Name), "s")
			code += "\tif len(request." + row.Name + ") > 0 {\n" +
				"\t\tquery = query.Where(\"" + usePrefixTable + "" + trimmed + " IN ?\", request." + row.Name + ")\n" +
				"\t}\n\n"
		case "bool":
			code += "\tquery = query.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " = ?\", request." + row.Name + ")\n" +
				"\n\n"
		case "[]byte":
			code += "// TODO: Настал то час когда нужно реализовать поиск фильтр по байтам" +
				"\n\n"
		case "*timestamp.Timestamp":
		case "*timestamppb.Timestamp":
			code += "\n\t// TODO Поставить правильное условие " + row.Name + "(" + strcase.ToSnake(row.Name) + ") in table " + nameInSnake + "\n"

			if row.Name == "DateStart" {
				code += "\tif request." + row.Name + " != nil {\n" +
					"\t\tquery.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " >= ?\", request." + row.Name + ".AsTime())\n" +
					"\t}\n\n"
				break
			}
			if row.Name == "DateEnd" {
				code += "\tif request." + row.Name + " != nil {\n" +
					"\t\tquery.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " <= ?\", request." + row.Name + ".AsTime())\n" +
					"\t}\n\n"
				break
			}

			if row.Name == "CreatedStart" {
				code += "\tif request." + row.Name + " != nil {\n" +
					"\t\tquery.Where(\"" + usePrefixTable + "created_at >= ?\", request." + row.Name + ".AsTime())\n" +
					"\t}\n\n"
				break
			}
			if row.Name == "CreatedEnd" {
				code += "\tif request." + row.Name + " != nil {\n" +
					"\t\tquery.Where(\"" + usePrefixTable + "created_at <= ?\", request." + row.Name + ".AsTime())\n" +
					"\t}\n\n"
				break
			}

			code += "\tif request." + row.Name + " != nil {\n" +
				"\t\tquery.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " = ?\", request." + row.Name + ".AsTime())\n" +
				"\t}\n\n"

		default:

			if strings.Contains(row.Name, "Type") || strings.Contains(row.Name, "Status") {
				code += "\tif request." + row.Name + " > 0 {\n" +
					"\t\tquery = query.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " = ?\", request." + row.Name + ")\n" +
					"\t}\n\n"
				break
			}

			code += "\n\n"
			code += "\t// TODO not implemented " + row.Name + "(" + strcase.ToSnake(row.Name) + ") in table " + nameInSnake
			code += "\n\n"
			log.Warn("Type: " + row.Name + "  " + row.Type + " not implemented (Generate Service ListFilter)")
		}
	}

	return
}
