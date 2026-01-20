package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/tools/go/loader"
)

const (
	skydivePkg = "atomgit.com/openeuler/PilotGo-plugin-topology"
)

var (
	filename string
	pkgname  string
	output   string
	strict   bool
	verbose  bool
	program  *loader.Program
	packages map[string]*types.Package
)

type module struct {
	Name      string
	BuildTags []string
	Getters   map[string]*getter
}

type field struct {
	Name    string
	Type    string
	IsArray bool
	Public  bool
}

func (f *field) ElemType() string {
	return strings.TrimLeft(f.Type, "[]")
}

type getter struct {
	Name    string
	IsArray bool
	Fields  []field
}

func (g *getter) Has(kind string) bool {
	for _, field := range g.Fields {
		if field.Type == kind {
			return true
		}
	}
	return false
}

func (g *getter) HasPublic() bool {
	for _, field := range g.Fields {
		if field.Public {
			return true
		}
	}
	return false
}

func (g *getter) HasArray() bool {
	for _, field := range g.Fields {
		if field.IsArray {
			return true
		}
	}
	return false
}

func resolveSymbol(pkg, symbol string) (types.Object, error) {
	if typePackage, found := packages[pkg]; found {
		return typePackage.Scope().Lookup(symbol), nil
	}

	return nil, fmt.Errorf("Failed to retrieve package info for %s", pkg)
}

func handleBasic(getter *getter, name, kind string) {
	switch kind {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		getter.Fields = append(getter.Fields, field{Name: name, Type: "int64"})
	default:
		public := false
		firstChar := strings.TrimPrefix(kind, "[]")
		if unicode.IsUpper(rune(firstChar[0])) {
			public = true
		}
		if splits := strings.SplitN(firstChar, ".", 2); len(splits) > 1 {
			firstChar = splits[1]
		}
		getter.Fields = append(getter.Fields, field{
			Name:    name,
			Type:    kind,
			IsArray: strings.HasPrefix(kind, "[]"),
			Public:  public,
		})
	}
}

func handleField(astFile *ast.File, getter *getter, fieldName, pkgName, fieldType string) error {
	switch fieldType {
	case "string", "bool", "int", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":
		handleBasic(getter, fieldName, fieldType)
	default:
		symbol, err := resolveSymbol(pkgName, fieldType)
		if err != nil {
			return fmt.Errorf("Failed to resolve symbol for %+v: %s", fieldType, err)
		}
		if symbol == nil {
			return fmt.Errorf("Failed to resolve symbol for %+v", fieldType)
		}

		if named, ok := symbol.Type().(*types.Named); ok {
			for i := 0; i < named.NumMethods(); i++ {
				if named.Method(i).Name() == "String" {
					handleBasic(getter, fieldName, "stringer")
					return nil
				}
			}
		}
		underlying := symbol.Type().Underlying()
		handleBasic(getter, fieldName, underlying.String())
	}

	return nil
}

func handleSpec(astFile *ast.File, getter *getter, spec interface{}) {
	if typeSpec, ok := spec.(*ast.TypeSpec); ok {
		if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		FIELD:
			for _, field := range structType.Fields.List {
				if len(field.Names) > 0 {
					fieldName := field.Names[0].Name
					var tag reflect.StructTag
					if field.Tag != nil {
						tag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					}

					if fieldType, found := tag.Lookup("field"); found {
						if fieldType == "omit" {
							continue FIELD
						}
						handleBasic(getter, fieldName, fieldType)
						continue
					} else if fieldType, ok := field.Type.(*ast.Ident); ok {
						if err := handleField(astFile, getter, fieldName, filepath.Base(pkgname), fieldType.Name); err != nil {
							log.Print(err)
						}
						continue
					} else if fieldType, ok := field.Type.(*ast.SelectorExpr); ok {
						if err := handleField(astFile, getter, fieldName, fieldType.X.(*ast.Ident).Name, fieldType.Sel.Name); err != nil {
							log.Print(err)
						}
						continue
					} else if fieldType, ok := field.Type.(*ast.ArrayType); ok {
						if starExpr, ok := fieldType.Elt.(*ast.StarExpr); ok {
							if itemIdent, ok := starExpr.X.(*ast.Ident); ok {
								handleBasic(getter, fieldName, "[]"+itemIdent.String())
								continue
							}
						} else if itemIdent, ok := fieldType.Elt.(*ast.Ident); ok {
							handleBasic(getter, fieldName, "[]"+itemIdent.String())
							continue
						}
					} else if fieldType, ok := field.Type.(*ast.StarExpr); ok {
						if itemIdent, ok := fieldType.X.(*ast.Ident); ok {
							handleBasic(getter, fieldName, itemIdent.String())
							continue
						} else if fieldType, ok := fieldType.X.(*ast.SelectorExpr); ok {
							handleBasic(getter, fieldName, fieldType.Sel.Name)
							continue
						}
					} else if fieldType, ok := field.Type.(*ast.MapType); ok {
						if keyIdent, ok := fieldType.Key.(*ast.Ident); ok {
							if valueIdent, ok := fieldType.Value.(*ast.Ident); ok {
								handleBasic(getter, fieldName, fmt.Sprintf("map[%s]%s", keyIdent.String(), valueIdent.String()))
								continue
							}
						}
					}

					if strict {
						log.Panicf("Don't know what to do with %s: %s", fieldName, spew.Sdump(field.Type))
					}
					if verbose {
						log.Printf("Don't know what to do with %s: %s", fieldName, spew.Sdump(field.Type))
					}
				} else {
					// Embedded field
					ident, _ := field.Type.(*ast.Ident)
					if starExpr, ok := field.Type.(*ast.StarExpr); ident == nil && ok {
						ident, _ = starExpr.X.(*ast.Ident)
					}

					if ident != nil {
						embedded := astFile.Scope.Lookup(ident.Name)
						if embedded != nil {
							handleSpec(astFile, getter, embedded.Decl)
						}
					}
				}
			}
		} else if arrayType, ok := typeSpec.Type.(*ast.ArrayType); ok {
			if starExpr, ok := arrayType.Elt.(*ast.StarExpr); ok {
				if itemIdent, ok := starExpr.X.(*ast.Ident); ok {
					getter.IsArray = true
					handleSpec(astFile, getter, itemIdent.Obj.Decl)
				}
			}
		} else {
			log.Printf("Don't know what to do with %s (%s)", typeSpec.Name, spew.Sdump(typeSpec))
		}
	}
}

func parseFile(filename string, pkgName string) (*module, error) {
	conf := loader.Config{
		ParserMode:  parser.ParseComments,
		AllowErrors: true,
		TypeChecker: types.Config{
			Error: func(err error) {
				if verbose {
					log.Print(err)
				}
			},
		},
	}
	astFile, err := conf.ParseFile(filename, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse %s: %s", filename, err)
	}

	conf.Import(pkgName)

	program, err = conf.Load()
	if err != nil {
		return nil, fmt.Errorf("Failed to load %s(%s): %s", filename, pkgName, err)
	}

	packages = make(map[string]*types.Package, len(program.AllPackages))
	for typePackage := range program.AllPackages {
		packages[typePackage.Name()] = typePackage
	}

	var buildTags []string
	for _, comment := range astFile.Comments {
		if strings.HasPrefix(comment.Text(), "+build ") {
			buildTags = append(buildTags, comment.Text())
		}
	}

	getters := make(map[string]*getter)
	for _, decl := range astFile.Decls {
		if decl, ok := decl.(*ast.GenDecl); ok {
			easyJSON := false
			if decl.Doc != nil {
				for _, doc := range decl.Doc.List {
					if easyJSON = strings.Index(doc.Text, "gendecoder") != -1; easyJSON {
						break
					}
				}
			}
			if !easyJSON {
				continue
			}

			getter := new(getter)
			for _, spec := range decl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					getter.Name = typeSpec.Name.Name
					handleSpec(astFile, getter, typeSpec)
				}
			}

			if getter.Name != "" {
				getters[getter.Name] = getter
			}
		}
	}

	return &module{
		Name:      astFile.Name.Name,
		BuildTags: buildTags,
		Getters:   getters,
	}, nil
}
