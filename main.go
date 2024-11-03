package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"log"
	"os"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	typeNames = flag.String("type", "", "comma-separated list of type names; must be set")
	output    = flag.String("output", "", "output file name; default srcdir/<type>_schema.go")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of govader:\n")
	fmt.Fprintf(os.Stderr, "\tgovader [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tgovader [flags] -type T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://pkg.go.dev/golang.org/x/tools/cmd/govader\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("govader: ")
	flag.Usage = Usage
	flag.Parse()
	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	typeNames := strings.Split(*typeNames, ",")

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()
	if len(args) == 0 {
		// Default: process whole package in current directory.
		args = []string{"."}
	}

	pkg, err := loadPackages(args)
	if err != nil {
		panic(err)
	}

	g := Generator{pkg: pkg}
	for _, typeName := range typeNames {
		values := findTypeValues(typeName, pkg)
		if len(values) > 0 {
			schema := ParseSchema(values)
			g.generate(schema)
		}
	}
}

func loadPackages(pattern []string) (*Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		return nil, err
	}
	gopkg := pkgs[0]
	pkg := &Package{Package: gopkg}
	var files []*File
	for _, f := range gopkg.Syntax {
		files = append(files, &File{
			pkg:  pkg,
			file: f,
		})
	}
	pkg.files = files
	return pkg, nil
}

func findTypeValues(typeName string, pkg *Package) []StructInfo {
	values := make([]StructInfo, 0, 100)
	for _, f := range pkg.files {
		f.typeName = typeName
		ast.Inspect(f.file, f.genDecl)
		values = append(values, f.values...)
	}
	return values
}

type Package struct {
	*packages.Package
	files []*File
}

type File struct {
	pkg      *Package
	file     *ast.File
	typeName string
	values   []StructInfo
}

func (f *File) genDecl(n ast.Node) bool {
	typeSpec, ok := n.(*ast.TypeSpec)
	if !ok {
		return true
	}

	structName := typeSpec.Name.Name
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok || structName != f.typeName {
		return true
	}

	value := StructInfo{
		name:      structName,
		fieldList: make([]FieldInfo, 0),
	}
	for _, field := range structType.Fields.List {
		for _, iden := range field.Names {
			fieldType := f.pkg.TypesInfo.TypeOf(field.Type)
			basicType := fieldType.Underlying().(*types.Basic).Info()

			var tag string
			if field.Tag != nil {
				tag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("gov")
			}
			value.fieldList = append(value.fieldList, FieldInfo{
				name: iden.Name,
				tag:  tag,
				typ:  basicType,
			})

		}
	}
	f.values = append(f.values, value)

	return false
}
