package entity

import "strings"

const TypeCurrent = 1
const TypeRequest = 2
const TypeResponse = 3

type Struct struct {
	Name string
	Rows []StructField
	Type int
}

type StructField struct {
	Name string
	Type string
	Tags string
}

type Tag struct {
	Name string
	Params []string
}

func GetTypeByName(name string) (typeStruct int) {

	if strings.HasSuffix(name,"Response") {
		return TypeResponse
	}

	if strings.HasSuffix(name,"Request") {
		return TypeRequest
	}

	if !strings.HasSuffix(name,"Server") && !strings.HasSuffix(name,"Client")   {
		return TypeCurrent
	}
	return

}