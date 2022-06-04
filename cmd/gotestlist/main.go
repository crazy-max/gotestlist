package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/build"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/crazy-max/gotestlist"
	gitignore "github.com/sabhiram/go-gitignore"
)

const usage = `usage: %s [-f=<format>] [-d=<size>] <packages>

gotestlist is looking for tests in the given list of packages.
It can also look for them recursively starting in the current directory by using: gotestlist ./...

Flags:
	-f: output format ; can be "json" or Go template layout (default "%s")
	-d: distribute tests based on the given matrix size ; output for each entry can be used with "go test -run (<matrix_entry>)/"
`

const defaultFormat = "{{.Pkg}}\t{{.Name}}\t{{.File}}"
const iterationTemplate = "{{range .}}%s\n{{end}}"

var format = flag.String("f", defaultFormat, "")
var distribute = flag.Int("d", 0, "")

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

	if *distribute > 0 {
		if err := runDistribute(ts, *distribute); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := runPrint(ts, *format); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runPrint(ts gotestlist.TestSlice, format string) error {
	w, f, fn := output(format)

	t, err := getTemplate(f)
	if err != nil {
		return err
	}

	if err = printTests(w, ts, f, t); err != nil {
		return err
	}

	if fn != nil {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func runDistribute(ts gotestlist.TestSlice, size int) error {
	suites := make(map[string]int)
	for _, t := range ts {
		name := t.Suite
		if name == "" {
			if !strings.HasSuffix(t.Name, "Suite") {
				name = t.Name
			} else {
				continue
			}
		}
		if _, ok := suites[name]; !ok {
			suites[name] = 0
		}
		suites[name]++
	}

	skeys := make([]string, 0, len(suites))
	for k := range suites {
		skeys = append(skeys, k)
	}
	sort.Strings(skeys)

	type matrixEntry struct {
		Suites []string
		Size   int
	}

	matrixEntries := make(map[int]*matrixEntry)
	msize := int(math.Ceil(float64(ts.Len()) / float64(size)))
	pos := 1
	for _, skey := range skeys {
		suiteName := skey
		suiteSize := suites[skey]
		if _, ok := matrixEntries[pos]; !ok {
			matrixEntries[pos] = &matrixEntry{}
		}
		if pos < size && matrixEntries[pos].Size > 0 && matrixEntries[pos].Size+suiteSize > msize {
			pos++
			if _, ok := matrixEntries[pos]; !ok {
				matrixEntries[pos] = &matrixEntry{}
			}
		}
		matrixEntries[pos].Size += suiteSize
		matrixEntries[pos].Suites = append(matrixEntries[pos].Suites, suiteName)
	}
	mkeys := make([]int, 0, len(matrixEntries))
	for k := range matrixEntries {
		mkeys = append(mkeys, k)
	}
	sort.Ints(mkeys)

	var matrix []string
	for _, mkey := range mkeys {
		matrix = append(matrix, strings.Join(matrixEntries[mkey].Suites, "|"))
	}

	b, _ := json.Marshal(matrix)
	fmt.Println(string(b))

	return nil
}
