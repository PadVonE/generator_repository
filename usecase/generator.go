package usecase

import (
	"generator/entity"
	"generator/generators"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func GenerateEntity(packageInfo entity.PackageStruct, serviceName string, listOfStruct []entity.Struct, replaceFile bool) {

	servicePath := filepath.FromSlash("./../" + serviceName)

	for _, l := range listOfStruct {
		if l.Type == entity.TypeMain {

			createFunction := false
			updateFunction := false

			// Определяем нужны ли функиции для создания и обновления даннх
			for _, tempStruct := range listOfStruct {
				if tempStruct.Name == "Create"+l.Name+"Request" {
					createFunction = true
				}
				if tempStruct.Name == "Update"+l.Name+"Request" {
					updateFunction = true
				}
			}

			code, err := generators.GenerateEntity(l, packageInfo, createFunction, updateFunction)
			if err != nil {
				log.Error(err)
				continue
			}

			saveFilePath := servicePath + "/entity/" + strcase.ToSnake(l.Name) + ".go"
			if replaceFile {
				err = FileSave(saveFilePath, code)
				if err == nil {
					log.WithField("File", saveFilePath).Println("Entity created")
				}
			}
		}
	}
}

func GenerateServiceFiles(packageInfo entity.PackageStruct, protoInterface entity.ProtoInterface, serviceName string, replaceFile bool) {
	var err error

	servicePath := filepath.FromSlash("./../" + serviceName)

	for _, pi := range protoInterface.Methods {
		// Generate file

		code := ""
		name, action := pi.NameInterface()

		saveFilePath := servicePath + "/service/" + strcase.ToSnake(name) + "_" + strcase.ToSnake(action) + ".go"

		// Если не удалось определить экшн то переходим к следующему методу
		if len(action) == 0 {
			continue
		}

		code, err = generators.GenerateServiceCode(pi, packageInfo, action)

		if err != nil {
			log.Error(err)
			continue
		}
		if replaceFile {
			err := FileSave(saveFilePath, code)

			if err == nil {
				log.WithField("File", saveFilePath).Println("Service file created ", strcase.ToSnake(name)+"_"+strcase.ToSnake(action)+".go")
			}
		}

	}
}

func GenerateTestFiles(packageInfo entity.PackageStruct, protoInterface entity.ProtoInterface, serviceName string, replaceFile bool) {
	var err error

	servicePath := filepath.FromSlash("./../" + serviceName)

	for _, pi := range protoInterface.Methods {
		// Generate file

		name, action := pi.NameInterface()

		// Если не удалось определить экшн то переходим к следующему методу
		if len(action) == 0 {
			continue
		}

		// Generate tests

		saveFileTestPath := servicePath + "/service/" + strcase.ToSnake(name) + "_" + strcase.ToSnake(action) + "_test.go"
		codeTest := ""
		switch action {
		case "Create":
			codeTest, err = generators.GenerateTestCreateCode(pi, packageInfo)
		case "Update":
			codeTest, err = generators.GenerateTestUpdateCode(pi, packageInfo)
		case "Delete":
			codeTest, err = generators.GenerateTestDeleteCode(pi, packageInfo)
		case "Get":
			codeTest, err = generators.GenerateTestGetCode(pi, packageInfo)
		case "List":
			codeTest, err = generators.GenerateTestListCode(pi, packageInfo)

		}

		if len(codeTest) != 0 {

			if err != nil {
				log.Error(err)
				continue
			}
			if replaceFile {
				err = FileSave(saveFileTestPath, codeTest)

				if err == nil {
					log.WithField("File", saveFileTestPath).Println("Test file created ", strcase.ToSnake(name)+"_"+strcase.ToSnake(action)+".go")
				}
			}
		}

	}
}

func GenerateMigrationFile(packageInfo entity.PackageStruct, serviceName string, listOfStruct []entity.Struct, replaceFile bool) {
	servicePath := filepath.FromSlash("./../" + serviceName)

	//Таблица для отслеживания изменений
	migration := "create table if not exists edited_log (\n" +
		"    id serial not null constraint edited_log_event_pkey primary key,\n" +
		"    created_at timestamp not null default CURRENT_TIMESTAMP,\n" +
		"    action text not null,\n" +
		"    table_name text not null,\n" +
		"    table_id integer not null,\n" +
		"    edited_user_id integer not null,\n" +
		"    json_string json not null\n);\n\n"

	// Инлекс
	migration += "create index if not exists edited_log_table_name_table_id_idx on edited_log (table_name, table_id);\n\n"

	// Тригер
	migration += "create or replace function edited_user_id() returns trigger\n" +
		"    language plpgsql\nas\n$$\nbegin\n" +
		"    if new.edited_user_id > 0 then\n" +
		"        insert into \"edited_log\" (\"action\", \"table_name\", \"table_id\", \"edited_user_id\", \"json_string\")\n" +
		"        values (tg_op, tg_table_name, new.id, new.edited_user_id, row_to_json(new.*));\n" +
		"    end if;\n" +
		"    return new;\nend;\n$$;\n\n"

	for _, l := range listOfStruct {
		if l.Type == entity.TypeMain {
			code, err := generators.GenerateMigration(l, packageInfo)
			if err != nil {
				log.Error(err)
				continue
			}
			migration += code
		}
	}
	now := time.Now()

	saveFilePath := servicePath + "/migrations/" + strconv.Itoa(int(now.Unix())) + "_init.up.sql"
	if replaceFile {
		err := FileSave(saveFilePath, migration)
		if err == nil {
			log.WithField("File", saveFilePath).Println("Migration created")
		}
	}
}

func GenerateGeneralFilesIfNotExist(packageInfo entity.PackageStruct, serviceName string, listOfStruct []entity.Struct, replaceFile bool) {

	type GeneralFile struct {
		FileName string
		Replace  bool
	}

	servicePath := filepath.FromSlash("./../" + serviceName)

	listFiles := []GeneralFile{
		{".gitignore", false},
		{"db.go", false},
		{"envopt.json", false},
		//{"envopt_test.json", false},
		{"go.mod", false},
		{"go.sum", false},
		{"main.go", false},
		{"server.go", false},
		{"service/service.go", false},
		//{"service/service_test.go", true},
		//{"prometheus.go",false},
		{"healthcheck/healthcheck.go", false},
	}

	dbList := []string{}
	for _, l := range listOfStruct {
		if l.Type == entity.TypeMain {
			dbList = append(dbList, strcase.ToSnake(l.Name))
		}
	}

	for _, l := range listFiles {
		saveFilePath := servicePath + "/" + l.FileName
		// Проверка на то что файл не существует
		if !l.Replace {
			if _, err := os.Stat(saveFilePath); err == nil {
				continue
			}
		}

		code, err := generators.GenerateGeneral(l.FileName, packageInfo, dbList)
		if err != nil {
			log.Error(err)
			continue
		}

		if replaceFile {
			err = FileSave(saveFilePath, code)
			if err == nil {
				log.WithField("File", saveFilePath).Println("GeneralFile created")
			}
		}

	}
}

func GenerateGatewayFiles(packageInfo entity.PackageStruct, protoInterface entity.ProtoInterface, serviceName string, replaceFile bool) {
	var err error

	servicePath := filepath.FromSlash("./../gateway-front-admin")

	for _, pi := range protoInterface.Methods {
		// Generate file

		code := ""
		name, action := pi.NameInterface()

		saveFilePath := servicePath + "/service/" + strcase.ToSnake(name) + "_" + strcase.ToSnake(action) + ".go"

		// Если не удалось определить экшн то переходим к следующему методу
		if len(action) == 0 {
			continue
		}

		code, err = generators.GenerateGatewayCode(pi, packageInfo, action)

		if err != nil {
			log.Error(err)
			continue
		}
		if replaceFile {
			err := FileSave(saveFilePath, code)

			if err == nil {
				log.WithField("File", saveFilePath).Println("Service file created ", strcase.ToSnake(name)+"_"+strcase.ToSnake(action)+".go")
			}
		}

	}
}

func GeneratePathProject(serviceName string) {
	servicePath := filepath.FromSlash("./../" + serviceName)
	pathList := []string{
		"entity",
		"migrations",
		"service",
		"healthcheck",
	}

	for _, path := range pathList {

		p := servicePath + "/" + path

		if _, err := os.Stat(p); os.IsNotExist(err) {
			os.Mkdir(p, os.ModePerm)
		}
	}

}
