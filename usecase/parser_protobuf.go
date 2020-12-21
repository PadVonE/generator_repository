package usecase

import (
	"bytes"
	"fmt"
	"generator/entity"
	log "github.com/sirupsen/logrus"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

const SourceComment = "// source: "
const SourceProtoPrefix = "crtm_"
const SourceProtoExtention = ".proto"

func ParseProtobufStruct(source string) (listOfStruct []entity.Struct) {

	listOfStruct = []entity.Struct{}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", []byte(source), 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	//typ := "" //для запоминания последнего определенного типа в ast
	//consts := make([]string, 0) //массив для сохранения найденных констант
	for _, decl := range f.Decls {
		//массив с определениями типов, переменных, констант, функций и т.п.

		switch n := decl.(type) {
		case *ast.GenDecl:
			switch n.Tok {
			case token.TYPE:

				ts := n.Specs[0].(*ast.TypeSpec)

				if tm, ok := ts.Type.(*ast.StructType); ok {
					rows := []entity.StructField{}
					for _, field := range tm.Fields.List {
						var typeNameBuf bytes.Buffer
						printer.Fprint(&typeNameBuf, fset, field.Type)

						if field.Tag == nil || len(field.Tag.Value) == 0 {
							continue
						}

						rows = append(rows, entity.StructField{
							Name: field.Names[0].Name,
							Type: typeNameBuf.String(),
							//Tags: field.Tag.Value,
						})

					}

					typeStruct := entity.GetTypeByName(ts.Name.Name)

					if typeStruct != 0 {
						listOfStruct = append(listOfStruct, entity.Struct{
							Name: ts.Name.Name,
							Rows: rows,
							Type: typeStruct,
						})
					} else {
						log.WithField("Struct", ts.Name.Name).Warn("Struct doesnt has type")
					}

				}

			}
		}

	}

	return
}

func ParseProtobufFunc(source string) (protoInterface entity.ProtoInterface) {
	protoInterface = entity.ProtoInterface{}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", []byte(source), 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, decl := range f.Decls {
		switch n := decl.(type) {
		case *ast.GenDecl:
			//log.Println(n.Tok)

			switch n.Tok {
			case token.TYPE:

				ts := n.Specs[0].(*ast.TypeSpec)

				if tt, ok := ts.Type.(*ast.InterfaceType); ok {

					methods := []entity.ProtoInterfaceMethod{}

					// Проходим по списку методов в интерфейсе
					for _, m := range tt.Methods.List {

						if tm, ok := m.Type.(*ast.FuncType); ok {
							//
							//results := []string{}
							//params := []string{}

							result := ""
							param := ""

							// Собираем список параметров ответа

							for _, item := range tm.Results.List {
								if ti, ok := item.Type.(*ast.StarExpr); ok {
									if d, ok := ti.X.(*ast.Ident); ok {
										result = d.Name
										break
										//results = append(results, d.Name)
									}
								}
							}

							// Собираем список параметров функции

							for _, item := range tm.Params.List {
								if ti, ok := item.Type.(*ast.StarExpr); ok {
									if d, ok := ti.X.(*ast.Ident); ok {
										param = d.Name
										break
										//params = append(params, d.Name)
									}
								}
							}

							basic := strings.Replace(param,"Request","",-1)
							basic = strings.Replace(basic,"Create","",-1)
							basic = strings.Replace(basic,"Update","",-1)
							basic = strings.Replace(basic,"Delete","",-1)
							basic = strings.Replace(basic,"Get","",-1)
							basic = strings.Replace(basic,"List","",-1)

							methods = append(methods, entity.ProtoInterfaceMethod{
								NameMethod: m.Names[0].Name,
								Request:    param,
								Response:   result,
								Basic:   	basic,
							})
						}

					}
					// Проверяем только интерфейсы которые преднозначины для клиента

					if strings.HasSuffix(ts.Name.Name, "Client") && len(tt.Methods.List) > 0 {
						protoInterface.NameInterface = ts.Name.Name
						protoInterface.Methods = methods
					}

				}

			}

		}

	}

	return
}

func ParseProtobufSourceAddress(source string) (pack entity.PackageStruct) {

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", []byte(source), parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	sourceAddress := ""
	for _, c := range f.Comments {

		for _, c1 := range c.List {
			contain := strings.Contains(c1.Text, SourceComment)
			if contain {
				sourceAddress = strings.Replace(c1.Text, SourceComment, "", -1)
			}
		}
	}

	address := strings.Split(sourceAddress, "/")

	if len(address) == 4 {
		PackageNameForImport := strings.TrimPrefix(address[2], "proto-")
		packageName := strings.Replace(PackageNameForImport, "-", "_", -1)

		pack = entity.PackageStruct{
			GitCompanyName:       address[1],
			GitRepositoryName:    address[2],
			PackageName:          packageName,
			PackageNameForImport: PackageNameForImport,
		}

	}

	return
}
