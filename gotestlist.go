package gotestlist

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Test describes a single test found in the *_test.go file
type Test struct {
	Name      string `json:"name"`
	Benchmark bool   `json:"benchmark"`
	Fuzz      bool   `json:"fuzz"`
	Suite     string `json:"suite"`
	Pkg       string `json:"pkg"`
	File      string `json:"file"`
}

// String returns a string representation of the Test
// in the form of 'package.Test filename'
func (t *Test) String() string {
	return fmt.Sprintf("%s %s %s", t.Pkg, t.Name, t.File)
}

// TestSlice attaches the methods of sort.Interface to []Test.
// Sorting in increasing order comparing package+testname.
type TestSlice []Test

func (s TestSlice) Len() int           { return len(s) }
func (s TestSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TestSlice) Less(i, j int) bool { return s[i].Pkg+s[i].Name < s[j].Pkg+s[j].Name }

// Sort is a convenience method.
func (s TestSlice) Sort() { sort.Sort(s) }

// Tests function searches for test function declarations in the given directory.
func Tests(dir string) (tests TestSlice, err error) {
	pkg, err := build.ImportDir(dir, build.ImportMode(0))
	if err != nil {
		if _, ok := err.(*build.NoGoError); !ok {
			return nil, err
		}
	}

	fset := token.NewFileSet()
	for _, filename := range append(pkg.TestGoFiles, pkg.XTestGoFiles...) {
		filename = filepath.Join(dir, filename)
		f, err := parser.ParseFile(fset, filename, nil, parser.Mode(0))
		if err != nil {
			return nil, err
		}
		for _, decl := range f.Decls {
			if fdecl, ok := decl.(*ast.FuncDecl); ok {
				if t := getTest(filename, f.Name.String(), fdecl); t != nil {
					tests = append(tests, *t)
				}
			}
		}
	}

	return tests, nil
}

func getTest(filename string, pkg string, fdecl *ast.FuncDecl) *Test {
	if fdecl.Type.Params == nil || len(fdecl.Type.Params.List) != 1 || len(fdecl.Type.Params.List[0].Names) > 1 {
		return nil
	}
	if fdecl.Type.Results != nil && len(fdecl.Type.Results.List) > 0 {
		return nil
	}

	test := &Test{
		Pkg:  pkg,
		File: filename,
	}

	paramType := types.ExprString(fdecl.Type.Params.List[0].Type)
	prefix := "Test"
	if strings.HasPrefix(fdecl.Name.String(), "Benchmark") && paramType == "*testing.B" {
		test.Benchmark = true
		prefix = "Benchmark"
	} else if strings.HasPrefix(fdecl.Name.String(), "Fuzz") && paramType == "*testing.F" {
		test.Fuzz = true
		prefix = "Fuzz"
	} else if !strings.HasPrefix(fdecl.Name.String(), "Test") || paramType != "*testing.T" {
		return nil
	}

	test.Name = fdecl.Name.String()
	if len(test.Name) > len(prefix) {
		r, _ := utf8.DecodeRuneInString(test.Name[len(prefix):])
		if unicode.IsLower(r) {
			return nil
		}
	}

	if fdecl.Recv != nil && len(fdecl.Recv.List) == 1 {
		recvType := types.ExprString(fdecl.Recv.List[0].Type)
		if strings.HasPrefix(recvType, "*") && strings.HasSuffix(recvType, "Suite") {
			test.Suite = strings.TrimPrefix(recvType, "*")
		} else {
			return nil
		}
	}

	return test
}
