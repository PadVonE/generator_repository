package entity

import (
	"github.com/iancoleman/strcase"
	"strings"
)

type ProtoInterface struct {
	NameInterface string
	Methods       []ProtoInterfaceMethod
}
type ProtoInterfaceMethod struct {
	NameMethod string

	Basic          string
	BasicStruct    Struct
	Request        string
	RequestStruct  Struct
	Response       string
	ResponseStruct Struct
}
type NameInterface struct {
	Name       string
	Action     string
	Additional string
}

func (pim *ProtoInterfaceMethod) NameInterface(protoInterface *ProtoInterface) (nameInterface NameInterface) {
	prefixList := []string{"List", "Get", "Create", "Update", "Delete"}

	nameInterface = NameInterface{}

	for _, prefix := range prefixList {
		if strings.HasPrefix(pim.NameMethod, prefix) {
			nameInterface.Action = prefix
			nameInterface.Name = strings.Replace(pim.NameMethod, prefix, "", 1)
		}
	}

	basicMethod := []string{}
	for _, pi := range protoInterface.Methods {
		if len(pi.BasicStruct.Rows) == 0 {
			continue // Если нет строк то переходим к следующему методу
		}
		basicMethod = append(basicMethod, pi.Basic)
	}

	basicMethod = removeDuplicates(basicMethod)

	for _, method := range basicMethod {
		if strings.HasPrefix(nameInterface.Name, method) {
			nameInterface.Additional = strings.Replace(nameInterface.Name, method, "", 1)
			nameInterface.Name = method
		}
	}

	return
}

func removeDuplicates(strings []string) []string {
	uniqueStrings := make(map[string]bool)
	result := make([]string, 0, len(strings))

	for _, str := range strings {
		if !uniqueStrings[str] {
			uniqueStrings[str] = true
			result = append(result, str)
		}
	}

	return result
}

func (ni *NameInterface) FileName() string {
	fileName := strcase.ToSnake(ni.Name) + "_" + strcase.ToSnake(ni.Action)
	if len(ni.Additional) > 0 {
		fileName = strcase.ToSnake(ni.Name) + "_" + strcase.ToSnake(ni.Action) + "_" + strcase.ToSnake(ni.Additional)
	}
	return fileName
}

func (ni *NameInterface) GetMethodName() string {
	return ni.Name + ni.Additional
}

func (ni *NameInterface) GetStructureName() string {
	return ni.Name
}
