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
	fixture := []struct {
		Args []string
		Dirs sort.StringSlice
	}{
		{
			Args: []string{"."},
			Dirs: sort.StringSlice{pkg.Dir},
		},
		{
			Args: []string{"./..."},
			Dirs: sort.StringSlice{pkg.Dir},
		},
		{
			Args: []string{"../gotestlist/"},
			Dirs: sort.StringSlice{pkg.Dir},
		},
		{
			Args: []string{"../../..."},
			Dirs: sort.StringSlice{
				pkg.Dir,
				filepath.Join(splitList[:len(splitList)-1]...),
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
		{
			Args: []string{".", "../../"},
			Dirs: sort.StringSlice{
				pkg.Dir,
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
		{
			Args: []string{pkgImport},
			Dirs: sort.StringSlice{
				pkg.Dir,
			},
		},
		{
			Args: []string{pkgRoot},
			Dirs: sort.StringSlice{
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
		{
			Args: []string{pkgRoot + "/..."},
			Dirs: sort.StringSlice{
				pkg.Dir,
				filepath.Join(splitList[:len(splitList)-1]...),
				filepath.Join(splitList[:len(splitList)-2]...),
			},
		},
	}
	for _, f := range fixture {
		dirs, err := testDirs(f.Args)
		if err != nil {
			t.Errorf("expected err=nil; got %v", err)
		}
		sortedDirs := make(sort.StringSlice, 0, len(dirs))
		for dir := range dirs {
			sortedDirs = append(sortedDirs, filepath.Clean(dir))
		}
		sortedDirs.Sort()
		f.Dirs.Sort()
		if !reflect.DeepEqual(sortedDirs, f.Dirs) {
			t.Errorf("expected %v; got %v", f.Dirs, sortedDirs)
		}
	}
}
