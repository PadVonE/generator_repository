package usecase

import (
	"generator/entity"
	"generator/generators/repository"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func GenerateEntity(packageInfo entity.PackageStruct, serviceName string, listOfStruct []entity.Struct, replaceFile bool) {
	log.Println("\033[35m", "\n\nEntity files", "\033[0m")

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
	log.Println("\033[35m", "\n\nService files", "\033[0m")

	var err error

	servicePath := filepath.FromSlash("./../" + serviceName)

	for _, pi := range protoInterface.Methods {
		// Generate file

		code := ""
		nameInterface := pi.NameInterface(&protoInterface)
		saveFilePath := servicePath + "/service/" + nameInterface.FileName() + ".go"

		// Если не удалось определить экшн то переходим к следующему методу
		if len(nameInterface.Action) == 0 {
			log.Warn("action not allowed:", pi.NameMethod)
			continue
		}

		code, err = generators.GenerateServiceCode(pi, packageInfo, nameInterface)

		if err != nil {
			log.Error(err)
			continue
		}
		if replaceFile {
			err := FileSave(saveFilePath, code)

			if err == nil {
				log.WithField("File", saveFilePath).Println("Service file created ", nameInterface.FileName()+".go")
			}
		}

	}
}

func GenerateTestFiles(packageInfo entity.PackageStruct, protoInterface entity.ProtoInterface, serviceName string, replaceFile bool) {
	log.Println("\033[35m", "\n\nTEST files", "\033[0m")

	var err error

	servicePath := filepath.FromSlash("./../" + serviceName)

	for _, pi := range protoInterface.Methods {
		// Generate file

		nameInterface := pi.NameInterface(&protoInterface)
		saveFileTestPath := servicePath + "/service/" + nameInterface.FileName() + "_test.go"

		// Если не удалось определить экшн то переходим к следующему методу
		if len(nameInterface.Action) == 0 {
			continue
		}

		// Generate tests

		codeTest := ""
		switch nameInterface.Action {
		case "Create":
			codeTest, err = generators.GenerateTestCreateCode(pi, packageInfo, nameInterface)
		case "Update":
			codeTest, err = generators.GenerateTestUpdateCode(pi, packageInfo, nameInterface)
		case "Delete":
			codeTest, err = generators.GenerateTestDeleteCode(pi, packageInfo, nameInterface)
		case "Get":
			codeTest, err = generators.GenerateTestGetCode(pi, packageInfo, nameInterface)
		case "List":
			codeTest, err = generators.GenerateTestListCode(pi, packageInfo, nameInterface)

		}

		if len(codeTest) != 0 {

			if err != nil {
				log.Error(err)
				continue
			}
			if replaceFile {
				err = FileSave(saveFileTestPath, codeTest)

				if err == nil {
					log.WithField("File", saveFileTestPath).Println("Test file created ", nameInterface.FileName()+".go")
				}
			}
		}

	}
}

func GenerateMigrationFile(packageInfo entity.PackageStruct, serviceName string, listOfStruct []entity.Struct, replaceFile bool) {
	log.Println("\033[35m", "\n\nMigration file", "\033[0m")

	servicePath := filepath.FromSlash("./../" + serviceName)
	migration := ""
	//Таблица для отслеживания изменений
	migrationEditLog := "create table if not exists edited_log(\n" +
		"    id serial not null constraint edited_log_event_pkey primary key,\n" +
		"    created_at timestamp not null default CURRENT_TIMESTAMP,\n" +
		"    action text not null,\n" +
		"    table_name text not null,\n" +
		"    table_id integer not null,\n" +
		"    edited_user_id integer not null,\n" +
		"    json_string json not null\n);\n\n"

	// Инлекс
	migrationEditLog += "create index if not exists edited_log_table_name_table_id_idx on edited_log (table_name, table_id);\n\n"

	// Тригер
	migrationEditLog += "create or replace function edited_user_id() returns trigger\n" +
		"    language plpgsql\nas\n$$\nbegin\n" +
		"    if new.edited_user_id > 0 then\n" +
		"        insert into \"edited_log\" (\"action\", \"table_name\", \"table_id\", \"edited_user_id\", \"json_string\")\n" +
		"        values (tg_op, tg_table_name, new.id, new.edited_user_id, row_to_json(new.*));\n" +
		"    end if;\n" +
		"    return new;\nend;\n$$;\n\n"

	hasEditedLog := false
	for _, l := range listOfStruct {
		if l.Type == entity.TypeMain {
			code, addEditLogTrigger, err := generators.GenerateMigration(l)
			if err != nil {
				log.Error(err)
				continue
			}

			if addEditLogTrigger == true {
				hasEditedLog = true
			}
			migration += code
		}
	}
	now := time.Now()

	if hasEditedLog {
		migration = migrationEditLog + migration
	}

	saveFilePath := servicePath + "/migrations/"
	saveFileName := strconv.Itoa(int(now.Unix())) + "_init.up.sql"

	// Проверим есть ли файл миграции

	entries, err := os.ReadDir(saveFilePath)
	if err != nil {
		log.Fatal(err)
	}

	migrationIsset := false
	for _, e := range entries {
		log.Println(e.Name())
		contain := strings.Contains(e.Name(), "_init")
		if contain {
			migrationIsset = true
			saveFileName = e.Name()
		}
	}

	if migrationIsset {
		log.Warn("Migration isset in path " + saveFilePath + saveFileName)
	}

	if replaceFile {
		err := FileSave(saveFilePath+saveFileName, migration)
		if err == nil {
			log.WithField("File", saveFilePath+saveFileName).Println("Migration created")
		}
	}
}

func GenerateGeneralFilesIfNotExist(packageInfo entity.PackageStruct, serviceName string, listOfStruct []entity.Struct, isGenerateTestFile bool, replaceFile bool) {
	log.Println("\033[35m", "\n\nGeneral Files file", "\033[0m")

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

	if isGenerateTestFile {
		listFiles = append(listFiles, GeneralFile{"service/service_test.go", true})
		listFiles = append(listFiles, GeneralFile{"envopt_test.json", false})
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
	log.Println("\033[35m", "\n\nGateway file", "\033[0m")

	var err error

	servicePath := filepath.FromSlash("./../gateway-front-admin")

	for _, pi := range protoInterface.Methods {
		// Generate file

		code := ""
		nameInterface := pi.NameInterface(&protoInterface)
		saveFilePath := servicePath + "/service/" + nameInterface.FileName() + "_test.go"

		// Если не удалось определить экшн то переходим к следующему методу
		if len(nameInterface.Action) == 0 {
			continue
		}

		//code, err = gateway_generator.GenerateGatewayCode(&entity.OperationInfo{}, packageInfo, nameInterface)

		if err != nil {
			log.Error(err)
			continue
		}
		if replaceFile {
			err := FileSave(saveFilePath, code)

			if err == nil {
				log.WithField("File", saveFilePath).Println("Service file created ", nameInterface.FileName()+".go")
			}
		}
	}
}

func GeneratePathProject(serviceName string) {

}
