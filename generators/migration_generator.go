package generators

import (
	"generator/entity"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
)

func GenerateMigration(strc entity.Struct, packageStruct entity.PackageStruct) (code string, err error) {

	code = "CREATE TABLE IF NOT EXISTS "+strcase.ToSnake(strc.Name)+"(\n"

	for i, s := range strc.Rows{
		code += "\t"
		if s.Name=="Id" {
			code += "id serial not null constraint "+strcase.ToSnake(strc.Name)+"_pkey primary key,\n"
			continue
		}

		switch s.Type {
			case "*timestamp.Timestamp":
				code += strcase.ToSnake(s.Name)+" timestamp not null default CURRENT_TIMESTAMP"
			case "int32":
				code += strcase.ToSnake(s.Name)+" integer not null default 0"
			case "string":
				code += strcase.ToSnake(s.Name)+" varchar(255) not null default ''"
			default:
				log.Warn("Type: " + s.Type + " not implemented (GenerateMigration)")


		}
		if (i+1)!=len(strc.Rows) {
			code += ",\n"
		}

	}
	code += "\n);\n\n\n"


	return
}
