package test

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestRootPackageExportedAPISnapshot(t *testing.T) {
	root := filepath.Clean("..")

	got, err := buildRootAPISnapshot(root)
	if err != nil {
		t.Fatalf("build root API snapshot: %v", err)
	}

	wantPath := filepath.Join("testdata", "root_api_snapshot.txt")
	wantBytes, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("read %s: %v", wantPath, err)
	}

	want := normalizeSnapshotNewlines(string(wantBytes))
	if got != want {
		t.Fatalf("root API snapshot mismatch for %s\n\n--- got ---\n%s", wantPath, got)
	}
}

func buildRootAPISnapshot(root string) (string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, root, func(info os.FileInfo) bool {
		name := info.Name()
		return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}, parser.SkipObjectResolution)
	if err != nil {
		return "", err
	}

	pkg, ok := pkgs["mux"]
	if !ok {
		return "", fmt.Errorf("package mux not found in %s", root)
	}

	typeLines := make([]string, 0)
	constLines := make([]string, 0)
	varLines := make([]string, 0)
	funcLines := make([]string, 0)
	methodLines := make([]string, 0)
	fieldLines := make([]string, 0)
	ifaceLines := make([]string, 0)

	fileNames := make([]string, 0, len(pkg.Files))
	for name := range pkg.Files {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)

	for _, name := range fileNames {
		file := pkg.Files[name]
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				switch d.Tok {
				case token.CONST:
					for _, spec := range d.Specs {
						valueSpec := spec.(*ast.ValueSpec)
						for _, ident := range valueSpec.Names {
							if ident.IsExported() {
								constLines = append(constLines, "const "+ident.Name)
							}
						}
					}
				case token.VAR:
					for _, spec := range d.Specs {
						valueSpec := spec.(*ast.ValueSpec)
						for _, ident := range valueSpec.Names {
							if ident.IsExported() {
								varLines = append(varLines, "var "+ident.Name)
							}
						}
					}
				case token.TYPE:
					for _, spec := range d.Specs {
						typeSpec := spec.(*ast.TypeSpec)
						if !typeSpec.Name.IsExported() {
							continue
						}
						typeName := formatTypeName(fset, typeSpec)
						typeLines = append(typeLines, formatTypeLine(fset, typeSpec, typeName))

						switch typeExpr := typeSpec.Type.(type) {
						case *ast.StructType:
							fieldLines = append(fieldLines, collectStructFieldLines(fset, typeName, typeExpr)...)
						case *ast.InterfaceType:
							ifaceLines = append(ifaceLines, collectInterfaceLines(fset, typeName, typeExpr)...)
						}
					}
				}
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				if d.Recv == nil {
					funcLines = append(funcLines, formatFuncDeclLine(fset, "func", "", d.Name.Name, d.Type))
					continue
				}
				recvType, ok := exportedReceiverType(fset, d.Recv)
				if !ok {
					continue
				}
				methodLines = append(methodLines, formatFuncDeclLine(fset, "method", recvType, d.Name.Name, d.Type))
			}
		}
	}

	sort.Strings(constLines)
	sort.Strings(varLines)
	sort.Strings(funcLines)
	sort.Strings(typeLines)
	sort.Strings(fieldLines)
	sort.Strings(ifaceLines)
	sort.Strings(methodLines)

	var sections []string
	sections = appendSection(sections, "const", constLines)
	sections = appendSection(sections, "var", varLines)
	sections = appendSection(sections, "func", funcLines)
	sections = appendSection(sections, "type", typeLines)
	sections = appendSection(sections, "field", fieldLines)
	sections = appendSection(sections, "iface", ifaceLines)
	sections = appendSection(sections, "method", methodLines)

	return strings.Join(sections, "\n\n") + "\n", nil
}

func appendSection(sections []string, name string, lines []string) []string {
	if len(lines) == 0 {
		return sections
	}
	return append(sections, "["+name+"]\n"+strings.Join(lines, "\n"))
}

func formatTypeName(fset *token.FileSet, typeSpec *ast.TypeSpec) string {
	if typeSpec.TypeParams == nil || len(typeSpec.TypeParams.List) == 0 {
		return typeSpec.Name.Name
	}

	params := make([]string, 0)
	for _, field := range typeSpec.TypeParams.List {
		constraint := formatNodeString(fset, field.Type)
		for _, name := range field.Names {
			params = append(params, name.Name+" "+constraint)
		}
	}

	return typeSpec.Name.Name + "[" + strings.Join(params, ", ") + "]"
}

func formatTypeLine(fset *token.FileSet, typeSpec *ast.TypeSpec, typeName string) string {
	if typeSpec.Assign.IsValid() {
		return "type " + typeName + " = " + formatNodeString(fset, typeSpec.Type)
	}

	switch typeSpec.Type.(type) {
	case *ast.StructType:
		return "type " + typeName + " struct"
	case *ast.InterfaceType:
		return "type " + typeName + " interface"
	default:
		return "type " + typeName + " " + formatNodeString(fset, typeSpec.Type)
	}
}

func collectStructFieldLines(fset *token.FileSet, typeName string, structType *ast.StructType) []string {
	if structType.Fields == nil {
		return nil
	}

	lines := make([]string, 0)
	for _, field := range structType.Fields.List {
		fieldType := formatNodeString(fset, field.Type)
		if len(field.Names) == 0 {
			name, ok := exportedEmbeddedName(field.Type)
			if ok {
				lines = append(lines, "field "+typeName+"."+name+" "+fieldType)
			}
			continue
		}

		for _, name := range field.Names {
			if name.IsExported() {
				lines = append(lines, "field "+typeName+"."+name.Name+" "+fieldType)
			}
		}
	}
	return lines
}

func collectInterfaceLines(fset *token.FileSet, typeName string, interfaceType *ast.InterfaceType) []string {
	if interfaceType.Methods == nil {
		return nil
	}

	lines := make([]string, 0)
	for _, field := range interfaceType.Methods.List {
		if len(field.Names) == 0 {
			lines = append(lines, "iface "+typeName+" embed "+formatNodeString(fset, field.Type))
			continue
		}

		for _, name := range field.Names {
			if !name.IsExported() {
				continue
			}
			funcType, ok := field.Type.(*ast.FuncType)
			if !ok {
				lines = append(lines, "iface "+typeName+"."+name.Name+" "+formatNodeString(fset, field.Type))
				continue
			}
			lines = append(lines, "iface "+typeName+"."+name.Name+formatFuncSignature(fset, funcType))
		}
	}
	return lines
}

func exportedReceiverType(fset *token.FileSet, recv *ast.FieldList) (string, bool) {
	if recv == nil || len(recv.List) == 0 {
		return "", false
	}
	if !isExportedTypeExpr(recv.List[0].Type) {
		return "", false
	}
	return formatNodeString(fset, recv.List[0].Type), true
}

func isExportedTypeExpr(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.IsExported()
	case *ast.StarExpr:
		return isExportedTypeExpr(x.X)
	case *ast.IndexExpr:
		return isExportedTypeExpr(x.X)
	case *ast.IndexListExpr:
		return isExportedTypeExpr(x.X)
	case *ast.SelectorExpr:
		return x.Sel.IsExported()
	default:
		return false
	}
}

func exportedEmbeddedName(expr ast.Expr) (string, bool) {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.IsExported() {
			return x.Name, true
		}
	case *ast.StarExpr:
		return exportedEmbeddedName(x.X)
	case *ast.IndexExpr:
		return exportedEmbeddedName(x.X)
	case *ast.IndexListExpr:
		return exportedEmbeddedName(x.X)
	case *ast.SelectorExpr:
		if x.Sel.IsExported() {
			return x.Sel.Name, true
		}
	}
	return "", false
}

func formatFuncDeclLine(fset *token.FileSet, kind, recvType, name string, funcType *ast.FuncType) string {
	line := kind + " "
	if recvType != "" {
		line += "(" + recvType + ") "
	}
	return line + name + formatFuncSignature(fset, funcType)
}

func formatFuncSignature(fset *token.FileSet, funcType *ast.FuncType) string {
	return "(" + formatFieldListTypes(fset, funcType.Params) + ")" + formatResultTypes(fset, funcType.Results)
}

func formatFieldListTypes(fset *token.FileSet, list *ast.FieldList) string {
	if list == nil || len(list.List) == 0 {
		return ""
	}

	parts := make([]string, 0)
	for _, field := range list.List {
		fieldType := formatNodeString(fset, field.Type)
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			parts = append(parts, fieldType)
		}
	}
	return strings.Join(parts, ", ")
}

func formatResultTypes(fset *token.FileSet, list *ast.FieldList) string {
	if list == nil || len(list.List) == 0 {
		return ""
	}

	types := formatFieldListTypes(fset, list)
	if strings.Count(types, ",") == 0 {
		return " " + types
	}
	return " (" + types + ")"
}

func formatNodeString(fset *token.FileSet, node any) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		panic(err)
	}
	return buf.String()
}

func normalizeSnapshotNewlines(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}
