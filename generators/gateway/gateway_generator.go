package generators

import (
	"bytes"
	"generator/entity"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type NamingConventions struct {
	GatewayMethodName              string
	GatewayMethodAction            string
	GatewayProjectNameInKebab      string
	GatewayTagNameInSnake          string
	RelatedProjectNameInPascal     string
	RelatedProjectNameInSnake      string
	RelatedProjectNameInKebab      string
	RelatedOrganisationNameInKebab string
}

type Data struct {
	Names            NamingConventions
	PackageName      string
	OperationTagName string
	ListFilter       string
	Imports          string
	ListRequest      string
	ListResponse     string
	Project          entity.Project
	Organization     entity.Organization
}

// Необходимы данные для генирации кода
// 1 Название метода
// 2 Action
// 3 Структура запроса для формирования фильтра
// 4 Структура ответа для генирации списка перменных
// 5 packageStruct на для вывода конкретного репозитория или юзкейса

func GenerateGatewayCode(oi *entity.OperationInfo, gatewayName string, gatewayAction string, projectName string, relatedOrganization entity.Organization, relatedProject entity.Project) (code string, err error) {

	path := filepath.FromSlash("./generators/gateway/template/gateway/_" + strings.ToLower(gatewayAction) + ".txt")

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

	listFilter, imports := gatewayListFilter(oi, "")
	listResponse, importsResponse := gatewayListResponse(oi, "")

	relatedProject.Name = strings.ReplaceAll(relatedProject.Name, "proto-", "")

	data := Data{
		Names: NamingConventions{
			GatewayMethodName:              gatewayName,
			GatewayMethodAction:            gatewayAction,
			GatewayProjectNameInKebab:      strcase.ToKebab(projectName),
			GatewayTagNameInSnake:          strcase.ToSnake(oi.Tag),
			RelatedProjectNameInPascal:     strcase.ToCamel(relatedProject.Name),
			RelatedProjectNameInSnake:      strcase.ToSnake(relatedProject.Name),
			RelatedProjectNameInKebab:      strcase.ToKebab(relatedProject.Name),
			RelatedOrganisationNameInKebab: relatedOrganization.Name,
		},
		//NameInSnake: strcase.ToSnake(gatewayName),
		//NameInCamel: strcase.ToLowerCamel(gatewayName),
		PackageName: projectName,
		//
		ListFilter:   listFilter,
		Project:      relatedProject,
		Organization: relatedOrganization,
		Imports:      imports + importsResponse,
		ListResponse: listResponse,
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		//return err
	}

	code = tpl.String()

	return
}

func gatewayListFilter(oi *entity.OperationInfo, nameInSnake string) (code string, imports string) {
	code = ""
	imports = ""

	//countImports := map[string]int{}
	//usePrefixTable := ""

	for _, row := range oi.Request {

		//Исключаем из фильтрации лимит и офсет
		if row.Name == "limit" || row.Name == "offset" {
			continue
		}

		nameWithBigID := strings.ReplaceAll(strcase.ToCamel(row.Name), "Id", "ID")

		switch row.Type {
		case "string":

			code +=
				"\tif params." + nameWithBigID + " != nil {\n" +
					"\t\trequest." + strcase.ToCamel(row.Name) + " = *params." + nameWithBigID + "\n" +
					"\t}\n\n"

		case "integer":

			code +=
				"\tif params." + nameWithBigID + " != nil {\n" +
					"\t\trequest." + strcase.ToCamel(row.Name) + " = *params." + nameWithBigID + "\n" +
					"\t}\n\n"
		case "array":

			code +=
				"\tif len(params." + nameWithBigID + ") > 0 {\n" +
					"\t\trequest." + strcase.ToCamel(row.Name) + " = *params." + nameWithBigID + "\n" +
					"\t}\n\n"
		default:

			//if strings.Contains(row.Name, "Type") || strings.Contains(row.Name, "Status") {
			//	code += "\tif request." + row.Name + " > 0 {\n" +
			//		"\t\tquery = query.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " = ?\", request." + row.Name + ")\n" +
			//		"\t}\n\n"
			//	break
			//}

			code += "\n\n"
			code += "\t// TODO not implemented " + row.Name + "(" + strcase.ToSnake(row.Name) + ") in table "
			code += "\n\n"
			log.Warn("Type: " + row.Name + "  " + row.Type + " not implemented (Generate Gateway ListFilter)")
		}
	}

	return
}

func gatewayListResponse(oi *entity.OperationInfo, nameInSnake string) (code string, imports string) {
	code = ""
	imports = ""

	//countImports := map[string]int{}
	//usePrefixTable := ""

	responseList := oi.Responses[200]

	for _, list := range responseList {
		if list.Name == "payload" {
			for _, row := range list.Children {
				if row.Name == "items" || row.Name == "item" {
					responseList = row.Children
				}
			}
		}
	}

	entity.SortProperties(responseList)

	for _, row := range responseList {

		//Исключаем из фильтрации лимит и офсет
		if row.Name == "limit" || row.Name == "offset" {
			continue
		}
		nameWithBigID := strcase.ToCamel(row.Name)
		nameWithBigID = strings.ReplaceAll(nameWithBigID, "Id", "ID")
		nameWithBigID = strings.ReplaceAll(nameWithBigID, "Uuid", "UUID")

		switch row.Type {
		case "string":

			switch row.Name {
			case "created_at", "updated_at", "deleted_at":
				code += nameWithBigID + ": strfmt.DateTime(item." + strcase.ToCamel(row.Name) + ".AsTime()),\n"
			default:
				if strings.Contains(row.Name, "date") {
					code += nameWithBigID + ": strfmt.DateTime(item." + strcase.ToCamel(row.Name) + ".AsTime()),\n"
					break
				}

				if strings.Contains(row.Name, "Uuid") {
					code += nameWithBigID + ": strfmt.UUID(item." + strcase.ToCamel(row.Name) + "),\n"
				}
				code += nameWithBigID + ": item." + strcase.ToCamel(row.Name) + ",\n"
			}

		case "integer":
			if strings.Contains(row.Name, "type") || strings.Contains(row.Name, "status") {
				code += nameWithBigID + ": int32(item." + strcase.ToCamel(row.Name) + "),\n"
				break
			}
			code += nameWithBigID + ": item." + strcase.ToCamel(row.Name) + ",\n"
		case "number":
			code += nameWithBigID + ": item." + strcase.ToCamel(row.Name) + ",\n"
		case "array":
			code += nameWithBigID + ": item." + strcase.ToCamel(row.Name) + ",\n"
		default:

			//if strings.Contains(row.Name, "Type") || strings.Contains(row.Name, "Status") {
			//	code += "\tif request." + row.Name + " > 0 {\n" +
			//		"\t\tquery = query.Where(\"" + usePrefixTable + "" + strcase.ToSnake(row.Name) + " = ?\", request." + row.Name + ")\n" +
			//		"\t}\n\n"
			//	break
			//}

			code += "\n\n"
			code += "\t// TODO not implemented " + row.Name + "(" + strcase.ToSnake(row.Name) + ") in table "
			code += "\n\n"
			log.Warn("Type: " + row.Name + "  " + row.Type + " not implemented (Generate Gateway ListResponse)")
		}
	}

	return
}
