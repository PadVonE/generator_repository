package generators

import (
	"fmt"
	"generator/entity"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
)

type Data struct {
	Name           string
	StructRows     string
	NameInSnake    string
	RepositoryName string
	ToProto        string
	CreateProtoTo  string
	UpdateProtoTo  string
	ListFilter  string
	PackageStruct  entity.PackageStruct
}

type DataTest struct {
	Name           string
	StructRows     string
	FilterBy       string
	Imports        string
	NameInSnake    string
	NameInPascale  string
	NameInCamel    string
	RepositoryName string
	PackageStruct  entity.PackageStruct

	FinishedStruct   string
	StructForRequest string
	TestList1        string
	TestList2        string
	Functions        string
}




func generateEqualList(s1 string, s2 string, p entity.Struct) (code string) {

	code = ""
	for _, element := range p.Rows {
		code += "\ts.Equal(" + s1 + "." + element.Name + ", " + s2 + "." + element.Name + ")"
		code += "\n"
	}

	return
}

func generateRowRequest(elementName string, elementType string, inc int) (codeEntity string, imports string) {

	codeEntity = ""
	switch elementType {
	case "string":
		imports += "\t\"github.com/google/uuid\"\n"
		codeEntity += "\t\t" + elementName + ":uuid.New().String(),\n"
	case "int32":
		codeEntity += "\t\t" + elementName + ":" + strconv.Itoa(inc+1) + ",\n"
	case "float64":
		codeEntity += "\t\t" + elementName + ":" + fmt.Sprintf("%f", rand.Float64()) + ",\n"
	case "bool":
		boolString := "false"
		if rand.Float32() < 0.5 {
			boolString = "true"
		}
		codeEntity += "\t\t" + elementName + ":" + boolString + ",\n"
	case "*timestamp.Timestamp":

		imports += "\t\"time\"\n"
		codeEntity += "\t\t" + elementName + ":time.Now(),\n"

	default:
		log.Warn("Type: " + elementType + " not implemented (generateRowRequest)")

	}

	return
}
