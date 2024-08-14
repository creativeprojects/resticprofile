package templates

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
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
//   - {{ "plain" | hex }} => "706c61696e"
//   - {{ "plain" | base64 }} => "cGxhaW4="
//   - {{ tempDir }} => "/path/to/unique-tempdir"
//   - {{ tempFile "filename" }} => "/path/to/unique-tempdir/filename"
func TemplateFuncs(funcs ...map[string]any) (templateFuncs map[string]any) {
	templateFuncs = map[string]any{
		"contains":        func(search any, src any) bool { return strings.Contains(toString(src), toString(search)) },
		"matches":         func(ptn string, src any) bool { return mustCompile(ptn).MatchString(toString(src)) },
		"replace":         func(old, new, src string) string { return strings.ReplaceAll(src, old, new) },
		"replaceR":        func(ptn, repl, src string) string { return mustCompile(ptn).ReplaceAllString(src, repl) },
		"lower":           strings.ToLower,
		"upper":           strings.ToUpper,
		"trim":            strings.TrimSpace,
		"trimPrefix":      func(prefix, src string) string { return strings.TrimPrefix(src, prefix) },
		"trimSuffix":      func(suffix, src string) string { return strings.TrimSuffix(src, suffix) },
		"split":           func(sep, src string) []any { return collect.From(strings.Split(src, sep), toAny[string]) },
		"splitR":          func(ptn, src string) []any { return collect.From(mustCompile(ptn).Split(src, -1), toAny[string]) },
		"join":            func(sep string, src []any) string { return strings.Join(collect.From(src, toString), sep) },
		"list":            func(args ...any) []any { return args },
		"map":             toMap,
		"base64":          func(src any) string { return base64.StdEncoding.EncodeToString([]byte(toString(src))) },
		"hex":             func(src any) string { return hex.EncodeToString([]byte(toString(src))) },
		"tempDir":         TempDir,
		"tempFile":        TempFile,
		"privateTempFile": MustPrivateTempFile,
		"env":             func() string { return TempFile(".env.none") }, // satisfies the {{env}} interface w.o. functionality
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

var compiled = sync.Map{}

func mustCompile(pattern string) *regexp.Regexp {
	value, ok := compiled.Load(pattern)
	if !ok {
		value, _ = compiled.LoadOrStore(pattern, regexp.MustCompile(pattern))
	}
	return value.(*regexp.Regexp)
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
	// sanitize filename
	name = mustCompile(`[^\w0-9_\-.]`).ReplaceAllString(name, "_")
	// create temp file
	filename = path.Join(TempDir(), name)
	if file, err := os.OpenFile(filename, os.O_CREATE, 0644); err == nil {
		_ = file.Close()
	} else if !os.IsExist(err) {
		panic(err)
	}
	return
}

// NotStrictlyPrivate indicates that a PrivateTempFile was successfully created but the OS reports that it can be accessed by others
var NotStrictlyPrivate = errors.New("the private temp file is not strictly accessible by owners only")

// PrivateTempFile is like TempFile but guarantees that the returned file can be accessed by owners only when err is nil
func PrivateTempFile(name string) (filename string, err error) {
	filename = TempFile(name)
	const privateMode = os.FileMode(0600)
	if err = os.Chmod(filename, privateMode); err == nil {
		if stat, e := os.Stat(filename); e != nil || stat.Mode() != privateMode {
			err = NotStrictlyPrivate
		}
	}
	return
}

// MustPrivateTempFile returns a strictly private temp file or panics if this is not supported (e.g. on Windows)
func MustPrivateTempFile(name string) string {
	if filename, err := PrivateTempFile(name); err == nil {
		return filename
	} else {
		panic(fmt.Errorf("failed creating private file %q (may be unsupported by this OS): %w", filename, err))
	}
}

// EnvFileReceiverFunc declares the backend interface for the "{{env}}" template function
type EnvFileReceiverFunc func() (profileKey string, receiveFile func(string))

// EnvFileFunc creates a template func to retrieve a profile .env file that can be used to pass variables between shells
func EnvFileFunc(receiverFunc EnvFileReceiverFunc) map[string]any {
	files := make(map[string]string)

	getEnvFile := func(profile string) (envFile string) {
		envFile = files[profile]
		if len(envFile) == 0 {
			var err error
			envFile, err = PrivateTempFile(fmt.Sprintf("%s.env", profile))
			if err != nil && !errors.Is(err, NotStrictlyPrivate) {
				panic(fmt.Errorf("failed setting permissions for %s: %w", envFile, err))
			}
			files[profile] = envFile
		}
		return
	}

	return map[string]any{
		"env": func() (envFile string) {
			profile, receive := receiverFunc()
			envFile = getEnvFile(profile)
			receive(envFile)
			return
		},
	}
}

func toAny[T any](arg T) any { return arg }

func toString(arg any) string {
	switch t := arg.(type) {
	case string:
		return t
	case []byte:
		return string(t)
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
