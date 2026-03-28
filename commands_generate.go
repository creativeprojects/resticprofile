package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/config/jsonschema"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/util/templates"
)

const pathTemplates = "contrib/templates"

//go:embed contrib/completion/bash-completion.sh
var bashCompletionScript string

//go:embed contrib/completion/zsh-completion.sh
var zshCompletionScript string

//go:embed contrib/completion/fish-completion.fish
var fishCompletionScript string

func generateCommand(output io.Writer, ctx commandContext) (err error) {
	args := ctx.request.arguments
	// enforce no-log
	logger := clog.GetDefaultLogger()
	handler := logger.GetHandler()
	logger.SetHandler(clog.NewDiscardHandler())

	if slices.Contains(args, "--bash-completion") {
		_, err = fmt.Fprintln(output, bashCompletionScript)
	} else if slices.Contains(args, "--config-reference") {
		err = generateConfigReference(output, args[slices.Index(args, "--config-reference")+1:])
	} else if slices.Contains(args, "--json-schema") {
		err = generateJsonSchema(output, args[slices.Index(args, "--json-schema")+1:])
	} else if slices.Contains(args, "--random-key") {
		ctx.flags.resticArgs = args[slices.Index(args, "--random-key"):]
		err = randomKey(output, ctx)
	} else if slices.Contains(args, "--zsh-completion") {
		_, err = fmt.Fprintln(output, zshCompletionScript)
	} else if slices.Contains(args, "--fish-completion") {
		_, err = fmt.Fprintln(output, fishCompletionScript)
	} else {
		err = fmt.Errorf("nothing to generate for: %s", strings.Join(args, ", "))
	}

	if err != nil {
		logger.SetHandler(handler)
	}
	return
}

//go:embed contrib/templates/*
var configReferenceTemplates embed.FS

func generateConfigReference(output io.Writer, args []string) error {
	resticVersion := restic.AnyVersion
	destination := "docs/content/reference"
	if slices.Contains(args, "--version") {
		args = args[slices.Index(args, "--version"):]
		if len(args) > 1 {
			resticVersion = args[1]
			args = args[2:]
		}
	}
	if slices.Contains(args, "--to") {
		args = args[slices.Index(args, "--to"):]
		if len(args) > 1 {
			destination = args[1]
			args = args[2:]
		}
	}

	data := config.NewTemplateInfoData(resticVersion)
	tpl := templates.New("config-reference", data.GetFuncs())
	templates, err := fs.Sub(configReferenceTemplates, pathTemplates)
	if err != nil {
		return fmt.Errorf("cannot load templates: %w", err)
	}

	if len(args) > 0 {
		tpl, err = tpl.ParseFiles(args...)
	} else {
		tpl, err = tpl.ParseFS(templates, "*.gomd")
	}

	if err != nil {
		return fmt.Errorf("parsing failed: %w", err)
	}

	staticPages := []struct {
		templateName string
		fileName     string
	}{
		{"reference.gomd", "_index.md"},
		{"global.gomd", "global.md"},
		{"profile.gomd", "profile/_index.md"},
		{"nested.gomd", "nested/_index.md"},
		{"groups.gomd", "groups.md"},
		{"value-types.gomd", "value-types.md"},
		{"json-schema.gomd", "json-schema.md"},
	}

	for _, staticPage := range staticPages {
		fmt.Fprintf(output, "generating %s...\n", staticPage.templateName)
		err = generateFileFromTemplate(tpl, data, filepath.Join(destination, staticPage.fileName), staticPage.templateName)
		if err != nil {
			return fmt.Errorf("unable to generate page %s: %w", staticPage.fileName, err)
		}
	}

	weight := 1
	for _, profileSection := range data.ProfileSections() {
		fmt.Fprintf(output, "generating profile section %s (weight %d)...\n", profileSection.Name(), weight)
		sectionData := SectionInfoData{
			DefaultData: data.DefaultData,
			Section:     profileSection,
			Weight:      weight,
		}
		err = generateFileFromTemplate(tpl, sectionData, filepath.Join(destination, "profile", profileSection.Name()+".md"), "profile.sub-section.gomd")
		if err != nil {
			return fmt.Errorf("unable to generate profile section %s: %w", profileSection.Name(), err)
		}
		weight++
	}

	weight = 1
	for _, nestedSection := range data.NestedSections() {
		fmt.Fprintf(output, "generating nested section %s (weight %d)...\n", nestedSection.Name(), weight)
		sectionData := SectionInfoData{
			DefaultData: data.DefaultData,
			Section:     nestedSection,
			Weight:      weight,
		}
		err = generateFileFromTemplate(tpl, sectionData, filepath.Join(destination, "nested", nestedSection.Name()+".md"), "profile.nested-section.gomd")
		if err != nil {
			return fmt.Errorf("unable to generate nested section %s: %w", nestedSection.Name(), err)
		}
		weight++
	}
	return nil
}

func generateFileFromTemplate(tpl *template.Template, data any, fileName, templateName string) error {
	err := os.MkdirAll(filepath.Dir(fileName), 0o755)
	if err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	err = tpl.ExecuteTemplate(file, templateName, data)
	if err != nil {
		return fmt.Errorf("cannot execute template: %w", err)
	}
	return nil
}

func generateJsonSchema(output io.Writer, args []string) (err error) {
	resticVersion := restic.AnyVersion
	if slices.Contains(args, "--version") {
		args = args[slices.Index(args, "--version"):]
		if len(args) > 1 {
			resticVersion = args[1]
			args = args[2:]
		}
	}

	if len(args) == 0 {
		return fmt.Errorf("missing type of json schema to generate (global, v1, v2)")
	}

	switch args[0] {
	case "global":
		data := config.NewTemplateInfoData(resticVersion)
		tpl := templates.New("", data.GetFuncs())
		templates, err := fs.Sub(configReferenceTemplates, pathTemplates)
		if err != nil {
			return fmt.Errorf("cannot load templates: %w", err)
		}
		tpl, err = tpl.ParseFS(templates, "config-schema.gojson")
		if err != nil {
			return fmt.Errorf("parsing failed: %w", err)
		}
		err = tpl.ExecuteTemplate(output, "config-schema.gojson", data)
		if err != nil {
			return fmt.Errorf("cannot execute template: %w", err)
		}
		return nil
	case "v1":
		return jsonschema.WriteJsonSchema(config.Version01, resticVersion, output)
	case "v2":
		return jsonschema.WriteJsonSchema(config.Version02, resticVersion, output)
	default:
		return fmt.Errorf("unknown json schema type: %s", args[0])
	}
}

// SectionInfoData is used as data for go templates that render profile section references
type SectionInfoData struct {
	templates.DefaultData

	Section config.SectionInfo
	Weight  int
}
