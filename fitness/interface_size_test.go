package fitness

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInterfaceSizes(t *testing.T) {
	maxMethods := 5
	projectRoot := ".."

	filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "fitness" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts := spec.(*ast.TypeSpec)
				iface, ok := ts.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				methodCount := len(iface.Methods.List)
				name := file.Name.Name + "." + ts.Name.Name

				if methodCount > maxMethods {
					t.Errorf(
						"%s (%s) has %d methods (max %d)",
						name, path, methodCount, maxMethods,
					)
				}
			}
		}
		return nil
	})
}
