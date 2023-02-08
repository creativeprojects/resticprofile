package templates

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/collect"
	"golang.org/x/exp/maps"
)

// TemplateFuncs declares a few standard functions to simplify working with templates
//
// Available functions:
//   - {{ "some old string" | replace "old" "new" }} => "some new string"
//   - {{ "some old string" | regex "(old)" "$1 and new" }} => "some old and new string"
//   - {{ "ABC" | lower }} => "abc"
//   - {{ "abc" | upper }} => "ABC"
//   - {{ "  A " | trim }} => "A"
//   - {{ "--A-" | trimPrefix " " }} => "A-"
//   - {{ "--A-" | trimSuffix " " }} => "--A"
//   - {{ "A,B,C" | split "," }} => ["A", "B", "C"]
//   - {{ "A,B,C" | split "," | join ";" }} => "A;B;C"
//   - {{ list "A" "B" "C" }} => ["A", "B", "C"]
//   - {{ tempDir }} => "/path/to/unique-tempdir"
//   - {{ tempFile "filename" }} => "/path/to/unique-tempdir/filename"
func TemplateFuncs(funcs ...map[string]any) (templateFuncs map[string]any) {
	toString := func(arg any) string { return fmt.Sprint(arg) }
	toAny := func(arg string) any { return arg }

	compiledRegex := sync.Map{}
	mustCompileRegex := func(pattern string) *regexp.Regexp {
		value, ok := compiledRegex.Load(pattern)
		if !ok {
			value, _ = compiledRegex.LoadOrStore(pattern, regexp.MustCompile(pattern))
		}
		return value.(*regexp.Regexp)
	}

	templateFuncs = map[string]any{
		"replace":    func(old, new, src string) string { return strings.ReplaceAll(src, old, new) },
		"regex":      func(ptn, repl, src string) string { return mustCompileRegex(ptn).ReplaceAllString(src, repl) },
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"trim":       strings.TrimSpace,
		"trimPrefix": func(prefix, src string) string { return strings.TrimPrefix(src, prefix) },
		"trimSuffix": func(suffix, src string) string { return strings.TrimSuffix(src, suffix) },
		"split":      func(sep, src string) []any { return collect.From(strings.Split(src, sep), toAny) },
		"join":       func(sep string, src []any) string { return strings.Join(collect.From(src, toString), sep) },
		"list":       func(args ...any) []any { return args },
		"tempDir":    TempDir,
		"tempFile":   TempFile,
	}

	for _, funcsMap := range funcs {
		maps.Copy(templateFuncs, funcsMap)
	}
	return
}

// New returns a new Template instance with configured funcs (including TemplateFuncs)
func New(name string, funcs ...map[string]any) (tpl *template.Template) {
	tpl = template.New(name)
	tpl.Funcs(TemplateFuncs(funcs...))
	return
}

var tempDirInitializer sync.Once

const tempDirName = "t"

// TempDir returns the volatile temporary directory that is returned by template function tempDir
func TempDir() string {
	dir, err := util.TempDir()
	if err == nil {
		dir = path.Join(filepath.ToSlash(dir), tempDirName) // must use slash, backslash is escape in some config files
		tempDirInitializer.Do(func() {
			err = os.MkdirAll(dir, 0755)
		})
	}
	if err != nil {
		panic(err)
	}
	return dir
}

// TempFile returns the volatile temporary file that is returned by template function tempFile
func TempFile(name string) (filename string) {
	filename = path.Join(TempDir(), name)
	if file, err := os.OpenFile(filename, os.O_CREATE, 0644); err == nil {
		_ = file.Close()
	} else if !os.IsExist(err) {
		panic(err)
	}
	return
}
