package usecase

import (
	"generator/entity"
	"generator/generators"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

func GenerateEntity(packageInfo entity.PackageStruct, serviceName string, listOfStruct []entity.Struct) {

	servicePath := filepath.FromSlash("./" + serviceName)

	for _, l := range listOfStruct {
		if l.Type == entity.TypeCurrent {
			code, err := generators.GenerateEntity(l, packageInfo)
			if err != nil {
				log.Error(err)
				continue
			}

			saveFilePath := servicePath + "/entity/" + strcase.ToSnake(l.Name) + ".go"

			err = FileSave(saveFilePath, code)
			if err == nil {
				log.WithField("File", saveFilePath).Println("Entity created")
			}
		}
	}
}

func GenerateServiceFiles(packageInfo entity.PackageStruct, protoInterface entity.ProtoInterface, serviceName string) {
	var err error

	servicePath := filepath.FromSlash("./" + serviceName)

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

		err := FileSave(saveFilePath, code)

		if err == nil {
			log.WithField("File", saveFilePath).Println("Service file created ", strcase.ToSnake(name)+"_"+strcase.ToSnake(action)+".go")
		}

	}
}


func GenerateTestFiles(packageInfo entity.PackageStruct, protoInterface entity.ProtoInterface, serviceName string) {
	var err error

	servicePath := filepath.FromSlash("./" + serviceName)

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

			err = FileSave(saveFileTestPath, codeTest)

			if err == nil {
				log.WithField("File", saveFileTestPath).Println("Test file created ", strcase.ToSnake(name)+"_"+strcase.ToSnake(action)+".go")
			}
		}

	}
}
