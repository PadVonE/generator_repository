package generators

import (
	"generator/entity"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"strings"
)

func GenerateMigration(strc entity.Struct) (code string, addEditLogTrigger bool, err error) {

	addEditLogTrigger = false

	code = "create table if not exists " + strcase.ToSnake(strc.Name) + " (\n"

	for i, s := range strc.Rows {
		code += "\t"
		if s.Name == "Id" {
			code += "id serial not null constraint " + strcase.ToSnake(strc.Name) + "_pkey primary key,\n"
			continue
		}

		if s.Name == "EditedUserId" {
			addEditLogTrigger = true
		}

		switch s.Type {
		case "*timestamp.Timestamp":
		case "*timestamppb.Timestamp":
			code += strcase.ToSnake(s.Name) + " timestamp not null default CURRENT_TIMESTAMP"
		case "int32":
			code += strcase.ToSnake(s.Name) + " integer not null default 0"
		case "string":
			code += strcase.ToSnake(s.Name) + " varchar not null default ''"
		case "float64":
			code += strcase.ToSnake(s.Name) + " numeric not null default 0"
		case "bool":
			code += strcase.ToSnake(s.Name) + "  boolean not null default false"
		case "[]string":
			code += strcase.ToSnake(s.Name) + "  varchar[] not null default '{}'"
		case "[]int32":
			code += strcase.ToSnake(s.Name) + "  integer[] not null default '{}'"
		default:
			if strings.Contains(s.Type, "Type") || strings.Contains(s.Type, "Status") {
				code += strcase.ToSnake(s.Name) + " integer not null default 0"
				break
			}

			log.Warn("Type: " + s.Type + " not implemented (GenerateMigration)")

		}
		if (i + 1) != len(strc.Rows) {
			code += ",\n"
		}

	}
	code += "\n);\n\n"
	if addEditLogTrigger {
		code += "create trigger " + strcase.ToSnake(strc.Name) + "_edited_user_id\n    after insert or update or delete\n    on " + strcase.ToSnake(strc.Name) + "\n    for each row\nexecute procedure edited_user_id();"
		code += "\n\n\n"
	}

	return
}
