package generators

import (
	"bytes"
	"generator/entity"
	"github.com/iancoleman/strcase"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func GenerateEntity(strc entity.Struct, packageStruct entity.PackageStruct) (code string, err error) {

	path := filepath.FromSlash("./generator/generators/template/_entity.txt")
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

	data := Data{
		Name:          strc.Name,
		NameInSnake:   strcase.ToSnake(strc.Name),
		StructRows:    rows,
		ToProto:       ToProto(strc, packageStruct.PackageName),
		CreateProtoTo: CreatrProtoTo(strc),
		UpdateProtoTo: UpdateProtoTo(strc),
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
	}

	return "\t" + field.Name + " " + fType + " " + tags + "\n"

}

func ToProto(strc entity.Struct, repositoryName string) (code string) {
	code = "\n"
	for _, row := range strc.Rows {
		switch row.Name {
		case "CreatedAt":
			code += "\tcreatedAt, _ := ptypes.TimestampProto(" + strcase.ToSnake(strc.Name) + ".CreatedAt)\n"
		case "UpdatedAt":
			code += "\tupdatedAt, _ := ptypes.TimestampProto(" + strcase.ToSnake(strc.Name) + ".UpdatedAt)\n"
		case "DeletedAt":
			code += "\tdeletedAt, _ := ptypes.TimestampProto(" + strcase.ToSnake(strc.Name) + ".DeletedAt)\n"
		case "PublicDate":
			code += "\tpublicDate, _ := ptypes.TimestampProto(" + strcase.ToSnake(strc.Name) + ".PublicDate)\n"
		}
	}

	code += "\n"
	code += "\treturn &" + repositoryName + "." + strc.Name + "{"

	for _, row := range strc.Rows {
		code += "\t\t"
		switch row.Name {
		case "CreatedAt":
			code += "CreatedAt:  createdAt,\n"
		case "UpdatedAt":
			code += "UpdatedAt:  updatedAt,\n"
		case "PublicDate":
			code += "PublicDate:  publicDate,\n"
		default:
			code += "" + row.Name + ":   " + strcase.ToSnake(strc.Name) + "." + row.Name + ",\n"
		}
	}
	code += "\t}"

	return
}

func CreatrProtoTo(strc entity.Struct) (code string) {

	code += "\treturn " + strc.Name + "{\n"
	for _, row := range strc.Rows {
		if row.Name == "Id" ||
			row.Name == "CreatedAt" ||
			row.Name == "UpdatedAt" ||
			row.Name == "DeletedAt" {
			continue
		}

		switch row.Type {
		case "*timestamp.Timestamp":
			code = "\t" + strcase.ToLowerCamel(row.Name) + ",_ := ptypes.Timestamp(proto." + row.Name + ")\n\n" + code
			code += "\t\t" + row.Name + ": " + strcase.ToLowerCamel(row.Name) + ",\n"
		default:
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"

		}

	}
	code += "\t}\n"

	return
}
func UpdateProtoTo(strc entity.Struct) (code string) {

	code += "\treturn " + strc.Name + "{\n"
	for _, row := range strc.Rows {
		if row.Name == "CreatedAt" ||
			row.Name == "UpdatedAt" ||
			row.Name == "DeletedAt" {
			continue
		}

		switch row.Type {
		case "*timestamp.Timestamp":
			code = "\t" + strcase.ToLowerCamel(row.Name) + ",_ := ptypes.Timestamp(proto." + row.Name + ")\n\n" + code
			code += "\t\t" + row.Name + ": " + strcase.ToLowerCamel(row.Name) + ",\n"
		default:
			code += "\t\t" + row.Name + ": proto." + row.Name + ",\n"

		}

	}
	code += "\t}\n"

	return
}
