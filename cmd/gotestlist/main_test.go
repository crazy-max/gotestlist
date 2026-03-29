package main

import (
	"go/build"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

const pkgImport = "github.com/crazy-max/gotestlist/cmd/gotestlist"
const pkgRoot = "github.com/crazy-max/gotestlist"

func TestDirs(t *testing.T) {
	gopath, rootDir, pkgDir := createDirFixture(t)
	t.Setenv("GO111MODULE", "off")
	t.Setenv("GOPATH", gopath)
	prevGOPATH := build.Default.GOPATH
	build.Default.GOPATH = gopath
	t.Cleanup(func() {
		build.Default.GOPATH = prevGOPATH
	})

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	cases := []struct {
		name string
		args []string
		dirs sort.StringSlice
	}{
		{
			name: "working dir",
			args: []string{"."},
			dirs: sort.StringSlice{pkgDir},
		},
		{
			name: "all dirs",
			args: []string{"./..."},
			dirs: sort.StringSlice{pkgDir},
		},
		{
			name: "relative dir",
			args: []string{"../gotestlist/"},
			dirs: sort.StringSlice{pkgDir},
		},
		{
			name: "relative all dirs",
			args: []string{"../../..."},
			dirs: sort.StringSlice{
				rootDir,
				filepath.Join(rootDir, "cmd"),
				pkgDir,
			},
		},
		{
			name: "working dir and relative dir",
			args: []string{".", "../../"},
			dirs: sort.StringSlice{
				pkgDir,
				rootDir,
			},
		},
		{
			name: "pkg import",
			args: []string{pkgImport},
			dirs: sort.StringSlice{
				pkgDir,
			},
		},
		{
			name: "pkg root",
			args: []string{pkgRoot},
			dirs: sort.StringSlice{
				rootDir,
			},
		},
		{
			name: "pkg root all",
			args: []string{pkgRoot + "/..."},
			dirs: sort.StringSlice{
				rootDir,
				filepath.Join(rootDir, "cmd"),
				pkgDir,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Chdir(pkgDir); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}

			dirs, err := testDirs(tt.args)
			if err != nil {
				t.Fatalf("expected err=nil; got %v", err)
			}

			sortedDirs := make(sort.StringSlice, 0, len(dirs))
			for dir := range dirs {
				sortedDirs = append(sortedDirs, filepath.Clean(dir))
			}
			sortedDirs.Sort()

			expected := append(sort.StringSlice(nil), tt.dirs...)
			expected.Sort()
			if !reflect.DeepEqual(sortedDirs, expected) {
				t.Fatalf("expected %v; got %v", expected, sortedDirs)
			}
		})
	}
}

func createDirFixture(t *testing.T) (string, string, string) {
	t.Helper()
	gopath := t.TempDir()
	rootDir := filepath.Join(gopath, "src", filepath.FromSlash(pkgRoot))
	pkgDir := filepath.Join(rootDir, "cmd", "gotestlist")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create fixture dir %q: %v", pkgDir, err)
	}
	return gopath, rootDir, pkgDir
}
