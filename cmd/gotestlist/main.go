package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/crazy-max/gotestlist"
	gitignore "github.com/sabhiram/go-gitignore"
)

const usage = `%s [-f=<format>] <packages>

Options:

	-f:
		it can be "json" or any other layout where

			{{.Name}} = test name
			{{.Benchmark}} = is benchmark
			{{.Fuzz}} = is fuzz
			{{.Suite}} = suite name
			{{.Pkg}}  = package
			{{.File}} = file path

		Default("%s")


gotestlist is looking for tests in the given list of packages.
It can also look for them recursively starting in the current directory by using: gotestlist ./...
`

const defaultFormat = "{{.Pkg}}\t{{.Name}}\t{{.File}}"
const iterationTemplate = "{{range .}}%s\n{{end}}"

var format = flag.String("f", defaultFormat, "")

type set map[string]struct{}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0], defaultFormat)
	}
}

func walkfunc(root string, dirs set) error {
	var gi *gitignore.GitIgnore
	if _, err := os.Stat(path.Join(root, ".gitignore")); err == nil {
		gi, err = gitignore.CompileIgnoreFile(path.Join(root, ".gitignore"))
		if err != nil {
			return err
		}
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".git") {
				return filepath.SkipDir
			}
			if gi != nil && gi.MatchesPath(info.Name()) {
				return filepath.SkipDir
			}
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			dirs[abs] = struct{}{}
		}
		return nil
	})
}

func recursiveArg(arg string) (string, bool) {
	if strings.HasSuffix(arg, "/...") {
		return arg[:len(arg)-4], true
	}
	return arg, false
}

func absDir(arg string) (string, error) {
	if strings.HasPrefix(arg, ".") {
		return filepath.Abs(arg)
	}
	p, err := build.Import(arg, "", build.FindOnly)
	if err != nil {
		return "", err
	}
	return p.Dir, nil
}

func testDirs(args []string) (set, error) {
	var dirs = make(set)
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		arg, rec := recursiveArg(arg)
		dir, err := absDir(arg)
		if err != nil {
			return nil, err
		}
		dirs[dir] = struct{}{}
		if rec {
			if err := walkfunc(dir, dirs); err != nil {
				return nil, err
			}
		}
	}
	return dirs, nil
}

func tests(dirs set) (ts gotestlist.TestSlice, err error) {
	for dir := range dirs {
		t, err := gotestlist.Tests(dir)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t...)
	}
	if len(ts) == 0 {
		return nil, errors.New("no tests were found")
	}
	ts.Sort()
	return ts, nil
}

func output(format string) (io.Writer, string, func() error) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 5, 0, 5, ' ', 0)
	return w, strings.Join(strings.Fields(format), "\t"), w.Flush
}

func getTemplate(format string) (*template.Template, error) {
	if format == "" {
		format = defaultFormat
	}
	return template.New("TestTemplate").Parse(fmt.Sprintf(iterationTemplate, format))
}

func printTests(w io.Writer, ts gotestlist.TestSlice, format string, t *template.Template) error {
	switch format {
	case "json":
		b, err := json.Marshal(ts)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, string(b)); err != nil {
			return err
		}
	default:
		return t.Execute(w, ts)
	}
	return nil
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	w, f, fn := output(*format)

	t, err := getTemplate(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dirs, err := testDirs(flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ts, err := tests(dirs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = printTests(w, ts, f, t); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if fn != nil {
		if err := fn(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
