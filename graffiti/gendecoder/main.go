package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
	"unicode"

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
