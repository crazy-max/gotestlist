package gotestlist

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"text/template"
)

const pkg = "github.com/crazy-max/gotestlist"

const testFileStr = "{{if .BuildTag}}{{.BuildTag}}\n\n{{else}}{{end}}package {{.Pkg}}\n\n" +
	"import \"testing\"\n\n" +
	"{{if .Type}}{{.Type}}\n\n{{else}}{{end}}" +
	"{{range .TestFuncs}}func {{.Recv}} {{.TestName}}({{.Params}}) {{.Results}} {{.Body}}\n{{end}}\n"

var testFileTemplate = template.Must(
	template.New("testFile").Parse(testFileStr))

var goos = []string{
	"android",
	"darwin",
	"dragonfly",
	"freebsd",
	"linux",
	"nacl",
	"netbsd",
	"openbsd",
	"plan9",
	"solaris",
	"windows",
}

type testFile struct {
	BuildTag  string
	Pkg       string
	Type      string
	TestFuncs []testFunc
}

type testFunc struct {
	Recv     string
	TestName string
	Params   string
	Results  string
	Body     string
}

func chooseOtherOS() string {
	for _, os := range goos {
		if runtime.GOOS != os {
			return os
		}
	}
	return ""
}

func removeTestFiles(files []string) error {
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func createTestFiles(t *testing.T, dir string) (files []string, err error) {
	filesData := []struct {
		File     string
		TestFile testFile
	}{
		{
			File: filepath.Join(dir, "package_test.go"),
			TestFile: testFile{
				Pkg: "gotestlist_test",
				TestFuncs: []testFunc{
					{
						TestName: "TestZPackage",
						Params:   "t *testing.T",
						Body:     "{}",
					},
				},
			},
		},
		{
			File: filepath.Join(dir, "tag_test.go"),
			TestFile: testFile{
				BuildTag: "// +build " + chooseOtherOS(),
				Pkg:      "gotestlist",
				TestFuncs: []testFunc{
					{
						TestName: "TestTag",
						Params:   "*testing.T",
						Body:     "{}",
					},
				},
			},
		},
		{
			File: filepath.Join(dir, "os_"+chooseOtherOS()+"_test.go"),
			TestFile: testFile{
				Pkg: "gotestlist",
				TestFuncs: []testFunc{
					{
						TestName: "TestDifferentOS",
						Params:   "*testing.T",
						Body:     "{}",
					},
				},
			},
		},
		{
			File: filepath.Join(dir, "func_test.go"),
			TestFile: testFile{
				Pkg:  "gotestlist",
				Type: "type foo int",
				TestFuncs: []testFunc{
					{
						TestName: "TestornotTest",
						Params:   "*testing.T",
						Body:     "{}",
					},
					{
						Recv:     "(f *foo)",
						TestName: "TestMethod",
						Params:   "*testing.T",
						Body:     "{}",
					},
					{
						Recv:     "(foo)",
						TestName: "TestMethod2",
						Params:   "*testing.T",
						Body:     "{}",
					},
					{
						TestName: "TestReturn",
						Params:   "*testing.T",
						Results:  "int",
						Body:     "{return 5}",
					},
					{
						TestName: "TestTwoParams",
						Params:   "t *testing.T, s string",
						Body:     "{}",
					},
					{
						TestName: "TestTwoParamsStringFirst",
						Params:   "s string, t *testing.T",
						Body:     "{}",
					},
					{
						TestName: "TestOneParamWrong",
						Params:   "s string",
						Body:     "{}",
					},
					{
						TestName: "Test",
						Params:   "*testing.T",
						Body:     "{}",
					},
					{
						TestName: "Test1",
						Params:   "*testing.T",
						Body:     "{}",
					},
					{
						TestName: "BenchmarkRandInt",
						Params:   "*testing.B",
						Body:     "{}",
					},
					{
						TestName: "FuzzHex",
						Params:   "*testing.F",
						Body:     "{}",
					},
					{
						Recv:     "(s *MySuite)",
						TestName: "TestBuild",
						Params:   "*testing.T",
						Body:     "{}",
					},
					{
						TestName: "NotATest",
						Params:   "*testing.T",
						Body:     "{}",
					},
				},
			},
		},
	}
	for _, data := range filesData {
		files = append(files, data.File)
		f, err := os.Create(data.File)
		if err != nil {
			return files, err
		}
		if err = testFileTemplate.Execute(f, data.TestFile); err != nil {
			return files, err
		}
		if err = f.Close(); err != nil {
			return files, err
		}
	}
	return files, err
}

func TestTests(t *testing.T) {
	p, err := build.Import(pkg, "", build.FindOnly)
	if err != nil {
		t.Fatal(err)
	}
	files, err := createTestFiles(t, p.Dir)
	if err != nil {
		t.Fatal(err)
	}
	defer removeTestFiles(files)
	expected := TestSlice{
		{
			Name:      "BenchmarkRandInt",
			Benchmark: true,
			File:      filepath.Join(p.Dir, "func_test.go"),
			Pkg:       "gotestlist",
		},
		{
			Name: "FuzzHex",
			Fuzz: true,
			File: filepath.Join(p.Dir, "func_test.go"),
			Pkg:  "gotestlist",
		},
		{
			Name: "Test",
			File: filepath.Join(p.Dir, "func_test.go"),
			Pkg:  "gotestlist",
		},
		{
			Name: "Test1",
			File: filepath.Join(p.Dir, "func_test.go"),
			Pkg:  "gotestlist",
		},
		{
			Name:  "TestBuild",
			Suite: "MySuite",
			File:  filepath.Join(p.Dir, "func_test.go"),
			Pkg:   "gotestlist",
		},
		{
			Name: "TestTests",
			File: filepath.Join(p.Dir, "gotestlist_test.go"),
			Pkg:  "gotestlist",
		},
		{
			Name: "TestZPackage",
			File: filepath.Join(p.Dir, "package_test.go"),
			Pkg:  "gotestlist_test",
		},
	}
	ts, err := Tests(p.Dir)
	if err != nil {
		t.Fatal(err)
	}
	expected.Sort()
	ts.Sort()
	if len(ts) != len(expected) {
		t.Errorf("expected len(ts) = %d; got %d", len(expected), len(ts))
	}
	if !reflect.DeepEqual(expected, ts) {
		t.Errorf("expected %v; got %v", expected, ts)
	}
}
