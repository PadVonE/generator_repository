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

func GenerateEntity(strc entity.Struct, packageStruct entity.PackageStruct,createFunc bool,updateFunc bool) (code string, err error) {

	path := filepath.FromSlash("./generators/template/_entity.txt")
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
	rows := ""

	for _, s := range strc.Rows {
		rows += generateRow(s)
	}

	createProtoTo := ""
	updateProtoTo := ""
	if createFunc {
		createProtoTo = CreateProtoTo(strc,packageStruct.PackageName)
	}
	if updateFunc {
		updateProtoTo = UpdateProtoTo(strc,packageStruct.PackageName)
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
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func generateRow(field entity.StructField) string {
	tags := "`xorm:\"" + strcase.ToSnake(field.Name) + "\"`"
	fType := field.Type

	switch field.Name {
	case "Id":
		tags = "`xorm:\"'id' pk autoincr\"`"
	case "CreatedAt":
		tags = "`xorm:\"created_at created\"`"
	case "UpdatedAt":
		tags = "`xorm:\"updated_at updated\"`"

	}

	switch field.Type {
	case "*timestamp.Timestamp":
		fType = "time.Time"
	default:
		if strings.Contains(field.Type, "Type") || strings.Contains(field.Type, "Status") {
			fType = "int32"
		}
	}

	return "\t" + field.Name + " " + fType + " " + tags + "\n"

}

func ToProto(strc entity.Struct, repositoryName string) (code string) {

	code += "\treturn &" + repositoryName + "." + strc.Name + "{"
	code += "\n"

	for _, row := range strc.Rows {
		code += "\t\t"
		switch row.Type {
		case "*timestamp.Timestamp":
			switch row.Name {
			case "CreatedAt":
				code += "CreatedAt:  timestamppb.New(" + strcase.ToLowerCamel(strc.Name) + ".CreatedAt),\n"
			case "UpdatedAt":
				code += "UpdatedAt:  timestamppb.New(" + strcase.ToLowerCamel(strc.Name) + ".UpdatedAt),\n"
			case "PublicDate":
				code += "PublicDate:  timestamppb.New(" + strcase.ToLowerCamel(strc.Name) + ".PublicDate),\n"
			}

		case "int32", "int64", "string", "float32", "float64":
			code += "" + row.Name + ":   " + strcase.ToLowerCamel(strc.Name) + "." + row.Name + ",\n"
		default:
			if strings.Contains(row.Type, "Type") || strings.Contains(row.Type, "Status")   {
				code += "" + row.Name + ": "+repositoryName+"."+row.Type+"(" + strcase.ToLowerCamel(strc.Name) + "."+row.Name + "),\n"
				break
			}

			log.Warn("Type: " + row.Type + " failed type (Generate Entity ToProto)")

		}

		//advertiser_repository.ContractType(contract.Type)
	}
	code += "\t}"

	return
}

func CreateProtoTo(strc entity.Struct,pkg string) (code string) {

	code += "func CreateProtoTo"+strc.Name+"(proto *"+pkg+".Create"+strc.Name+"Request) "+strc.Name+" {"

	code += "\n\treturn " + strc.Name + "{\n"
	for _, row := range strc.Rows {
		if row.Name == "Id" ||
			row.Name == "CreatedAt" ||
			row.Name == "UpdatedAt" ||
			row.Name == "DeletedAt" {
			continue
		}

		switch row.Type {
		case "*timestamp.Timestamp":
			code = "\t" + strcase.ToLowerCamel(row.Name) + " := timestamppb.New(proto." + row.Name + ")\n\n" + code
			code += "\t\t" + row.Name + ": " + strcase.ToLowerCamel(row.Name) + ",\n"
		case "int32", "int64", "string", "float32", "float64":
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

	return
}

func UpdateProtoTo(strc entity.Struct,pkg string) (code string) {

	code += "func UpdateProtoTo"+strc.Name+"(proto *"+pkg+".Update"+strc.Name+"Request) "+strc.Name+" {"

	code += "\n\treturn " + strc.Name + "{\n"
	for _, row := range strc.Rows {
		if row.Name == "CreatedAt" ||
			row.Name == "UpdatedAt" ||
			row.Name == "DeletedAt" {
			continue
		}

		switch row.Type {
		case "*timestamp.Timestamp":
			code = "\t" + strcase.ToLowerCamel(row.Name) + ",_ := timestamppb.New(proto." + row.Name + ")\n\n" + code
			code += "\t\t" + row.Name + ": " + strcase.ToLowerCamel(row.Name) + ",\n"
		case "int32", "int64", "string", "float32", "float64":
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

	return
}
