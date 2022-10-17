package main

import (
	"generator/entity"
	"generator/usecase"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {

	// Клонируем репозиторий
	clonePath := CloneRepository()

	// Парсим информацию из прото файлов в нужные нам структуры
	packageInfo, listOfStruct, funcList := ParseInfoFromProto(clonePath)

	// Групируем Структуры с методами в которых они вызываются
	GroupStructWithMethod(&funcList, listOfStruct)

	if IsCreateProjectStructure {
		// Создание структуры папок
		usecase.GeneratePathProject(ServiceName)
	}

	// Генерируем файлы со структурами
	if IsGenerateEntity {
		usecase.GenerateEntity(packageInfo, ServiceName, listOfStruct, ReplaceFile)
	}

	// Генерируем файлы Миграции
	if IsGenerateMigrationFile {
		usecase.GenerateMigrationFile(packageInfo, ServiceName, listOfStruct, ReplaceFile)
	}

	// Генерируем файлы реализаций методов
	if IsGenerateServiceFile {
		usecase.GenerateServiceFiles(packageInfo, funcList, ServiceName, ReplaceFile)
	}

	// Генерируем файлы тестов
	if IsGenerateTestFile {
		usecase.GenerateTestFiles(packageInfo, funcList, ServiceName, ReplaceFile)
	}

	// Генерируем файлы тестов
	if IsGenerateGatewayFile {
		usecase.GenerateGatewayFiles(packageInfo, funcList, ServiceName, ReplaceFile)
	}

	usecase.GenerateGeneralFilesIfNotExist(packageInfo, ServiceName, listOfStruct, ReplaceFile)

	// Выравнивание сгенеренного кода
	servicePath := filepath.FromSlash("./../" + ServiceName + "/")
	cmd := exec.Command("gofmt", "-s", "-w", servicePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

}

func CloneRepository() (clonePath string) {
	clonePath = usecase.CloningRepository(GitRepository,
		&http.BasicAuth{
			Username: GitHubUsername,
			Password: GitHubToken,
		})

	return
}

// packageInfo - Хранит информацию о названии пакета, полного пути на гите, названия для импорти,
// listOfStruct - список всех структур описанных в протофайлах (Название структуры,Тип структуры, Поля)
// funcList - Структура, (в которой распарсеный файл *_repository.proto), Хранит список методов для реализации а также request и response структуры

func ParseInfoFromProto(clonePath string) (packageInfo entity.PackageStruct, listOfStruct []entity.Struct, funcList entity.ProtoInterface) {

	files, err := ioutil.ReadDir(clonePath)
	if err != nil {
		log.Fatal(err)
	}

	funcFile := ""
	protoFiles := []string{}

	for _, file := range files {

		fileAddress := clonePath + "/" + file.Name()
		if strings.HasSuffix(file.Name(), "repository.pb.go") {
			funcFile = fileAddress
			continue
		}

		if strings.HasSuffix(file.Name(), ".pb.go") {
			protoFiles = append(protoFiles, fileAddress)
		}
	}

	packageInfo = usecase.GetRepositoryInfo(funcFile)

	listOfStruct = []entity.Struct{}

	for _, file := range protoFiles {

		dat, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		sourceFile := string(dat)
		listOfStruct = append(listOfStruct, usecase.ParseProtobufStruct(sourceFile)...)
	}

	dat, err := ioutil.ReadFile(funcFile)
	if err != nil {
		panic(err)
	}
	source := string(dat)

	funcList = usecase.ParseProtobufFunc(source)

	return
}

func GroupStructWithMethod(funcList *entity.ProtoInterface, listOfStruct []entity.Struct) {

	for i, f := range funcList.Methods {

		for _, los := range listOfStruct {
			if f.Request == los.Name {
				funcList.Methods[i].RequestStruct = los
			}
			if f.Response == los.Name {
				funcList.Methods[i].ResponseStruct = los
			}

			if f.Basic == los.Name {
				funcList.Methods[i].BasicStruct = los
			}
		}
	}
}
