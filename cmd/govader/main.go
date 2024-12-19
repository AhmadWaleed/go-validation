package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

var (
	typeNames = flag.String("type", "", "comma-separated list of type names; must be set")
	output    = flag.String("output", "", "output file name; default srcdir/<type>_schema.go")
	locale    = flag.String("locale", "en", "locale to use for error messages; default en")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of govader:\n")
	fmt.Fprintf(os.Stderr, "\tgovader [flags] -type T [directory]\n")
	fmt.Fprintf(os.Stderr, "\tgovader [flags] -type T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://github.com/AhmadWaleed/go-validation\n")
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

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."} // Default to current directory.
	}
	dir := filepath.Dir(args[0])

	pkg, err := loadPackage(args)
	if err != nil {
		panic(err)
	}

	var foundTypes []string
	var typeInfo []StructInfo
	for _, typeName := range typeNames {
		values := findTypeValues(typeName, pkg)
		if len(values) > 0 {
			foundTypes = append(foundTypes, typeName)
			typeInfo = append(typeInfo, values...)
		}
	}

	if len(typeInfo) == 0 {
		log.Println("no types found")
		return
	}

	schemas, err := parseSchema(typeInfo)
	if err != nil {
		log.Fatalf("invalid schema: %s", err)
	}

	buf := new(bytes.Buffer) // Accumulated output.
	g := &Generator{
		w:              buf,
		Schemas:        schemas,
		Messages:       LoadLocale(*locale),
		GeneratedRules: make(map[string]bool),
	}
	tmpl := &Template{
		PackageName: pkg.Package.Name,
		Generator:   g,
	}
	tmpl.Render(context.Background(), buf)

	// Format the output.
	src := gofmt(buf)

	// Write to file.
	outputName := *output
	if outputName == "" {
		outputName = filepath.Join(dir, baseName(foundTypes[0]))
	}
	err = os.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

func loadPackage(pattern []string) (*Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, pattern...)
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
	values := make([]StructInfo, 0, 10)
	for _, f := range pkg.files {
		f.typeName = typeName
		ast.Inspect(f.file, f.scanTypeStruct)
		values = append(values, f.values...)
		f.values = nil
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

func (f *File) scanTypeStruct(n ast.Node) bool {
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
		Name:      structName,
		FieldList: make([]FieldInfo, 0),
	}
	for _, field := range structType.Fields.List {
		for _, iden := range field.Names {
			fieldType := f.pkg.TypesInfo.TypeOf(field.Type)
			basicType := fieldType.Underlying().(*types.Basic).Kind()

			var tag string
			if field.Tag != nil {
				tag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1]).Get("gov")
			}
			if tag == "" || tag == "-" {
				continue
			}
			value.FieldList = append(value.FieldList, FieldInfo{
				Name: iden.Name,
				Tag:  tag,
				Type: basicType,
			})
		}
	}
	f.values = append(f.values, value)

	return false
}

// gofmt formats and returns the gofmt-ed contents of given buffer.
func gofmt(buf *bytes.Buffer) []byte {
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return buf.Bytes()
	}
	return src
}

func baseName(typename string) string {
	suffix := "schema.go"
	return fmt.Sprintf("%s_%s", strings.ToLower(typename), suffix)
}
