package main

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const pkgImport = "github.com/crazy-max/gotestlist/cmd/gotestlist"
const pkgRoot = "github.com/crazy-max/gotestlist"

var pkg *build.Package

func init() {
	var err error
	pkg, err = build.Import(pkgImport, "", build.FindOnly)
	if err != nil {
		panic(err)
	}
}

func TestDirs(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("getwd failed: ", err)
	}
	if err := os.Chdir(pkg.Dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)
	splitList := strings.SplitAfter(pkg.Dir, string(filepath.Separator))
	cases := []struct {
		name string
		args []string
		dirs sort.StringSlice
	}{
		{
			name: "working dir",
			args: []string{"."},
			dirs: sort.StringSlice{pkg.Dir},
		},
		{
			name: "all dirs",
			args: []string{"./..."},
			dirs: sort.StringSlice{pkg.Dir},
		},
		{
			name: "relative dir",
			args: []string{"../gotestlist/"},
			dirs: sort.StringSlice{pkg.Dir},
		},
		{
			name: "relative all dirs",
			args: []string{"../../..."},
			dirs: sort.StringSlice{
				pkg.Dir,
				filepath.Join(splitList[:len(splitList)-1]...),
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
		{
			name: "working dir and relative dir",
			args: []string{".", "../../"},
			dirs: sort.StringSlice{
				pkg.Dir,
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
		{
			name: "pkg import",
			args: []string{pkgImport},
			dirs: sort.StringSlice{
				pkg.Dir,
			},
		},
		{
			name: "pkg root",
			args: []string{pkgRoot},
			dirs: sort.StringSlice{
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
		{
			name: "pkg root all",
			args: []string{pkgRoot + "/..."},
			dirs: sort.StringSlice{
				pkg.Dir,
				filepath.Join(splitList[:len(splitList)-1]...),
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dirs, err := testDirs(tt.args)
			if err != nil {
				t.Errorf("expected err=nil; got %v", err)
			}
			sortedDirs := make(sort.StringSlice, 0, len(dirs))
			for dir := range dirs {
				sortedDirs = append(sortedDirs, filepath.Clean(dir))
			}
			sortedDirs.Sort()
			tt.dirs.Sort()
			if !reflect.DeepEqual(sortedDirs, tt.dirs) {
				t.Errorf("expected %v; got %v", tt.dirs, sortedDirs)
			}
		})
	}
}
