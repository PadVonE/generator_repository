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

func GenerateEntity(strc entity.Struct, packageStruct entity.PackageStruct, createFunc bool, updateFunc bool) (code string, err error) {

	path := filepath.FromSlash("./generators/repository/template/_entity.txt")
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
	rows, imports := generateRow(strc)

	createProtoTo := ""
	updateProtoTo := ""
	if createFunc {
		createProtoTo = CreateProtoTo(strc, packageStruct.PackageName)
	}
	if updateFunc {
		updateProtoTo = UpdateProtoTo(strc, packageStruct.PackageName)
	}

	data := Data{
		Name:          strc.Name,
		NameInSnake:   strcase.ToSnake(strc.Name),
		NameInCamel:   strcase.ToLowerCamel(strc.Name),
		StructRows:    rows,
		ToProto:       ToProto(strc, packageStruct.PackageName),
		CreateProtoTo: createProtoTo,
		UpdateProtoTo: updateProtoTo,
		PackageStruct: packageStruct,
		Imports:       imports,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func generateRow(row entity.Struct) (code string, imports string) {
	tags := ""
	code = ""
	imports = ""
	countImports := map[string]int{}

	for _, field := range row.Rows {

		fType := field.Type

		switch field.Name {
		case "Id":
			tags = ""
		case "CreatedAt":
			tags = "`gorm:\"->;<-:create\"`"
		case "UpdatedAt":
			tags = ""

		}
		//	GeoCodes            pq.StringArray `gorm:"type:char(2)[]"`
		switch field.Type {
		case "*timestamp.Timestamp":
		case "*timestamppb.Timestamp":
			if field.Name == "DeletedAt" {
				if countImports["gorm"] == 0 {
					imports += "\t\"gorm.io/gorm\"\n"
				}
				countImports["gorm"]++
				fType = "gorm.DeletedAt"
				break
			}

			fType = "time.Time"
		case "[]string":
			if countImports["[]string"] == 0 {
				imports += "\t\"github.com/lib/pq\"\n"
			}
			countImports["[]string"]++

			fType = "pq.StringArray"
			tags = "`gorm:\"type:varchar[]\"`"
		default:
			if strings.Contains(field.Type, "Type") || strings.Contains(field.Type, "Status") {
				fType = "int32"
			}
		}
		code += "\t" + field.Name + " " + fType + " " + tags + "\n"
	}
	return
}

func ToProto(strc entity.Struct, repositoryName string) (code string) {

	code += "\treturn &" + repositoryName + "." + strc.Name + "{"
	code += "\n"

	for _, row := range strc.Rows {
		code += "\t\t"
		switch row.Type {
		case "*timestamp.Timestamp", "*timestamppb.Timestamp":
			switch row.Name {
			case "DeletedAt":
				code += "DeletedAt:  timestamppb.New(" + strcase.ToLowerCamel(strc.Name) + ".DeletedAt.Time),\n"
				continue
			}

			code += "" + row.Name + ":  timestamppb.New(" + strcase.ToLowerCamel(strc.Name) + "." + row.Name + "),\n"

		case "int32", "int64":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		case "float32", "float64":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		case "string":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		case "bool":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		case "[]string":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		case "[]int32":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		case "[]byte":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		default:
			if strings.Contains(row.Type, "Type") || strings.Contains(row.Type, "Status") {
				code += "" + row.Name + ": " + repositoryName + "." + row.Type + "(" + strcase.ToLowerCamel(strc.Name) + "." + row.Name + "),\n"
				break
			}

			log.Warn("Type: " + row.Type + " failed type (Generate Entity ToProto)")

		}

		//advertiser_repository.ContractType(contract.Type)
	}
	code += "\t}"

	return
}

func CreateProtoTo(strc entity.Struct, pkg string) (code string) {

	startFinction := "func CreateProtoTo" + strc.Name + "(proto *" + pkg + ".Create" + strc.Name + "Request) " + strc.Name + " {"

	code += "\n\treturn " + strc.Name + "{\n"
	for _, row := range strc.Rows {
		if row.Name == "Id" ||
			row.Name == "CreatedAt" ||
			row.Name == "UpdatedAt" ||
			row.Name == "DeletedAt" {
			continue
		}

		switch row.Type {
		case "*timestamp.Timestamp", "*timestamppb.Timestamp":
			//	code += "\t" + strcase.ToLowerCamel(row.Name) + " := timestamppb.New(proto." + row.Name + ")\n\n" + code
			code += "\t\t" + row.Name + ":  proto." + row.Name + ".AsTime(),\n"
		case "int32", "int64":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "float32", "float64":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "string":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "bool":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "[]string":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "[]byte":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "[]int32":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		default:
			if strings.Contains(row.Type, "Type") || strings.Contains(row.Type, "Status") {
				code += "\t\t" + row.Name + ": int32(proto." + row.Name + "),\n"
				break
			}

			log.Warn("Type: " + row.Type + " failed type (Generate Entity CreateProtoTo)")
		}

	}
	code += "\t}\n"
	code += "}"

	return startFinction + code
}

func UpdateProtoTo(strc entity.Struct, pkg string) (code string) {

	startFinction := "func UpdateProtoTo" + strc.Name + "(proto *" + pkg + ".Update" + strc.Name + "Request) " + strc.Name + " {"

	variableInFunction := ""

	code += "\n\treturn " + strc.Name + "{\n"
	for _, row := range strc.Rows {
		if row.Name == "CreatedAt" ||
			row.Name == "UpdatedAt" ||
			row.Name == "DeletedAt" {
			continue
		}

		switch row.Type {
		case "*timestamp.Timestamp", "*timestamppb.Timestamp":
			//variableInFunction += "\n\t" + strcase.ToLowerCamel(row.Name) + ",_ := proto." + row.Name + ".asTime()\n\n"
			code += "\t\t" + row.Name + ": proto." + row.Name + ".AsTime(),\n"
		case "int32", "int64":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "float32", "float64":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "string":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "bool":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "[]string":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "[]byte":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		case "[]int32":
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"
		default:
			if strings.Contains(row.Type, "Type") || strings.Contains(row.Type, "Status") {
				code += "\t\t" + row.Name + ": int32(proto." + row.Name + "),\n"
				break
			}

			log.Warn("Type: " + row.Type + " failed type (Generate Entity UpdateProtoTo)")

		}

	}
	code += "\t}\n"
	code += "}"

	return startFinction + variableInFunction + code
}
