package templates

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
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
//   - {{ "some string" | contains "some" }} => true
//   - {{ "some string" | matches "^.+str.+$" }} => true
//   - {{ "some old string" | replace "old" "new" }} => "some new string"
//   - {{ "some old string" | replaceR "(old)" "$1 and new" }} => "some old and new string"
//   - {{ "some old string" | regex "(old)" "$1 and new" }} => "some old and new string" (alias to "replaceR")
//   - {{ "ABC" | lower }} => "abc"
//   - {{ "abc" | upper }} => "ABC"
//   - {{ "  A " | trim }} => "A"
//   - {{ "--A-" | trimPrefix " " }} => "A-"
//   - {{ "--A-" | trimSuffix " " }} => "--A"
//   - {{ "A,B,C" | split "," }} => ["A", "B", "C"]
//   - {{ "A,B,C" | split "," | join ";" }} => "A;B;C"
//   - {{ "A ,B, C" | splitR "\\s*,\\s*" | join ";" }} => "A;B;C"
//   - {{ list "A" "B" "C" }} => ["A", "B", "C"]
//   - {{ with $v := map "k1" "v1" "k2" "v2" }} {{ .k1 }}-{{ .k2 }} {{ end }}  => " v1-v2 "
//   - {{ with $v := list "A" "B" "C" "D" | map }} {{ ._0 }}-{{ ._1 }}-{{ ._3 }} {{ end }}  => " A-B-D "
//   - {{ with $v := list "A" "B" "C" "D" | map "key" }} {{ .key | join "-" }} {{ end }}  => " A-B-C-D "
//   - {{ tempDir }} => "/path/to/unique-tempdir"
//   - {{ tempFile "filename" }} => "/path/to/unique-tempdir/filename"
func TemplateFuncs(funcs ...map[string]any) (templateFuncs map[string]any) {
	compiled := sync.Map{}
	mustCompile := func(pattern string) *regexp.Regexp {
		value, ok := compiled.Load(pattern)
		if !ok {
			value, _ = compiled.LoadOrStore(pattern, regexp.MustCompile(pattern))
		}
		return value.(*regexp.Regexp)
	}

	templateFuncs = map[string]any{
		"contains":   func(search any, src any) bool { return strings.Contains(toString(src), toString(search)) },
		"matches":    func(ptn string, src any) bool { return mustCompile(ptn).MatchString(toString(src)) },
		"replace":    func(old, new, src string) string { return strings.ReplaceAll(src, old, new) },
		"replaceR":   func(ptn, repl, src string) string { return mustCompile(ptn).ReplaceAllString(src, repl) },
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"trim":       strings.TrimSpace,
		"trimPrefix": func(prefix, src string) string { return strings.TrimPrefix(src, prefix) },
		"trimSuffix": func(suffix, src string) string { return strings.TrimSuffix(src, suffix) },
		"split":      func(sep, src string) []any { return collect.From(strings.Split(src, sep), toAny[string]) },
		"splitR":     func(ptn, src string) []any { return collect.From(mustCompile(ptn).Split(src, -1), toAny[string]) },
		"join":       func(sep string, src []any) string { return strings.Join(collect.From(src, toString), sep) },
		"list":       func(args ...any) []any { return args },
		"map":        toMap,
		"tempDir":    TempDir,
		"tempFile":   TempFile,
	}

	// aliases
	templateFuncs["regex"] = templateFuncs["replaceR"]

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

func toAny[T any](arg T) any { return arg }

func toString(arg any) string {
	switch t := arg.(type) {
	case string:
		return t
	case []any:
		return "[" + strings.Join(collect.From(t, toString), ",") + "]"
	default:
		return fmt.Sprint(arg)
	}
}

func toMap(args ...any) (m map[string]any) {
	m = make(map[string]any)
	var key *string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			if key == nil {
				key = &v
			} else {
				m[*key] = v
				key = nil
			}
		default:
			if key != nil {
				m[*key] = v
			} else if v != nil && reflect.TypeOf(v).Kind() == reflect.Slice {
				rv := reflect.ValueOf(v)
				for i, length := 0, rv.Len(); i < length; i++ {
					if value := rv.Index(i); value.CanInterface() {
						k := fmt.Sprintf("_%d", i)
						m[k] = value.Interface()
					}
				}
			}
			key = nil
		}
	}
	return
}
