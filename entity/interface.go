package entity

import "strings"

type ProtoInterface struct {
	NameInterface string
	Methods []ProtoInterfaceMethod
}
type ProtoInterfaceMethod struct {
	NameMethod string

	Basic string
	BasicStruct Struct
	Request string
	RequestStruct Struct
	Response string
	ResponseStruct Struct
}

func (pim *ProtoInterfaceMethod) NameInterface() (name string,action string)  {
	prefixList := []string{"List","Get","Create","Update","Delete"}

	for _,prefix :=range prefixList {
		if strings.HasPrefix(pim.NameMethod,prefix) {
			action = prefix
			name =  strings.Replace(pim.NameMethod, prefix, "", 1)
		}
	}

	return
}