package jsonschema

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/util"
	"golang.org/x/exp/slices"
)

const (
	// for best compatibility the schema is mostly "draft-07", but some new vocabulary of 2019-09 is used
	jsonSchemaVersion    = "https://json-schema.org/draft/2019-09/schema"
	schemaUrlTemplate    = "/resticprofile/jsonschema/config-%d%s.json"
	referenceUrlTemplate = "#/$defs/%s"
	arrayTitleSuffix     = "..."
)

type schemaRoot struct {
	Schema       string                `json:"$schema"`
	Id           string                `json:"$id"`
	Defs         map[string]SchemaType `json:"$defs,omitempty"`
	schemaObject                       // cannot use SchemaType here as long as ",inline" support is missing in json.Marshal
}

func newSchema(version config.Version, id string, content *schemaObject) (root *schemaRoot, err error) {
	walkTypes(content, func(item SchemaType) SchemaType {
		if err == nil {
			if err = item.verify(); err != nil {
				bytes, _ := json.Marshal(item)
				if len(bytes) > 1024 {
					bytes = bytes[0:1024]
				}
				err = fmt.Errorf("verify: %s\nerror: %w", string(bytes), err)
			}
			return item
		}
		return nil
	})

	if err == nil {
		if len(id) > 0 {
			id = fmt.Sprintf("-restic-%s", strings.ReplaceAll(id, ".", "-"))
		}
		root = &schemaRoot{
			Schema:       jsonSchemaVersion,
			Id:           fmt.Sprintf(schemaUrlTemplate, version, id),
			Defs:         make(map[string]SchemaType),
			schemaObject: *content,
		}
	}

	return
}

// createReferences replaces duplicate definitions by references to shared definitions
func (r *schemaRoot) createReferences(minContentLength int) {
	names := make(map[SchemaType]string)

	// 3 passes (1: ref-count, 2: array-ref-count, 3: all type-lists)
	for i := 0; i < 3; i++ {
		refs := make(map[SchemaType]int)

		switch i {
		case 0: // Count references pass 1
			walkTypes(&r.schemaObject, func(item SchemaType) SchemaType {
				switch item.(type) {
				case *schemaTypeList, *schemaArray:
				default:
					refs[item]++
				}
				return item
			})
		case 1: // Count references pass 2
			walkTypes(&r.schemaObject, func(item SchemaType) SchemaType {
				switch item.(type) {
				case *schemaTypeList, *schemaReference:
				default:
					refs[item]++
				}
				return item
			})
		default: // Force sharing of schemaTypeList in last pass
			walkTypes(&r.schemaObject, func(item SchemaType) SchemaType {
				if _, typeList := item.(*schemaTypeList); typeList {
					refs[item] = 2
				}
				return item
			})
		}

		// Replace reused schemas with references
		walkTypes(&r.schemaObject, func(item SchemaType) SchemaType {
			if refs[item] > 1 {
				name := ""
				if defName, ok := names[item]; ok {
					name = defName
				} else {
					if bytes, err := json.Marshal(item); err == nil && len(bytes) > minContentLength {
						hash := sha256.Sum256(bytes) // use content hash to avoid duplicates
						name = fmt.Sprintf("id-%s", hex.EncodeToString(hash[0:16]))
						r.Defs[name] = item
						names[item] = name
					}
				}
				if len(name) > 0 {
					return &schemaReference{Ref: fmt.Sprintf(referenceUrlTemplate, name)}
				}
			}
			return item
		})
	}
}

// SchemaType is the interface for all json schema types
type SchemaType interface {
	base() *schemaTypeBase
	verify() error
	describe(title, description string)
}

// schemaTypeBase implements SchemaType and serves as base for most schemaType variants
type schemaTypeBase struct {
	Title       string `json:"title,omitempty"`
	Default     any    `json:"default,omitempty"`
	Deprecated  *bool  `json:"deprecated,omitempty"`
	Description string `json:"description,omitempty"`
	// Type value MUST be either a string or an array. If it is an array, elements of
	// the array MUST be strings and MUST be unique. String values MUST be one of the
	// types ("null", "boolean", "object", "array", "number", "integer", or "string").
	Type any `json:"type"`
	// Enum validates successfully if a value is equal to one of the elements in the array.
	Enum     []any `json:"enum,omitempty"`
	Examples []any `json:"examples,omitempty"`
}

func (s *schemaTypeBase) base() *schemaTypeBase {
	return s
}

func (s *schemaTypeBase) describe(title, description string) {
	s.Title = title
	s.Description = description
}

func (s *schemaTypeBase) verify() (err error) {
	// verify s.Type
	{
		var typeNames []string
		switch t := s.Type.(type) {
		case string:
			typeNames = []string{t}
		case []string:
			typeNames = t
		}
		if len(typeNames) == 0 {
			err = fmt.Errorf("expected single type or list of types, but none was specified")
		}
		for _, name := range typeNames {
			if !slices.Contains(validTypeNames, name) {
				err = fmt.Errorf("type name %q not in (%s)", name, strings.Join(validTypeNames, ", "))
			}
		}
	}
	return
}

func (s *schemaTypeBase) setDeprecated(value bool) {
	if value {
		s.Deprecated = util.CopyRef(value)
	} else {
		s.Deprecated = nil
	}
}

var validTypeNames = []string{"null", "boolean", "object", "array", "number", "integer", "string"}

func withBaseType[T SchemaType](t T, typeNames ...string) T {
	if base := t.base(); base != nil {
		if len(typeNames) == 1 {
			base.Type = typeNames[0]
		} else if len(typeNames) > 1 {
			base.Type = typeNames
		}
	}
	return t
}

// schemaTypeWithoutBase implements SchemaType without providing the base fields (for special cases)
type schemaTypeWithoutBase struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func (s *schemaTypeWithoutBase) base() *schemaTypeBase { return nil }
func (s *schemaTypeWithoutBase) verify() error         { return nil }

func (s *schemaTypeWithoutBase) describe(title, description string) {
	s.Title = title
	s.Description = description
}

type schemaTypeList struct {
	schemaTypeWithoutBase
	OneOf []SchemaType `json:"oneOf,omitempty"`
	AnyOf []SchemaType `json:"anyOf,omitempty"`
}

func (s *schemaTypeList) verify() error {
	if len(s.AnyOf) == 0 && len(s.OneOf) == 0 {
		return fmt.Errorf("neither anyOf nor oneOf defined")
	} else if len(s.AnyOf) > 0 && len(s.OneOf) > 0 {
		return fmt.Errorf("both, anyOf and oneOf defined")
	}

	// normalizing Description & Title
	for _, items := range [][]SchemaType{s.AnyOf, s.OneOf} {
		for _, item := range items {
			if base := item.base(); base != nil {
				if len(s.Description) == 0 && len(base.Description) > 0 {
					s.Description = base.Description
				}
				if len(s.Title) == 0 && len(base.Title) > 0 {
					s.Title = strings.TrimSuffix(base.Title, arrayTitleSuffix)
				}
				if s.Title == base.Title {
					base.Title = ""
				}
				if s.Description == base.Description {
					base.Description = ""
				}
			}
		}
	}
	return nil
}

func newSchemaTypeList(anyType bool, types ...SchemaType) *schemaTypeList {
	if anyType {
		return &schemaTypeList{AnyOf: types}
	} else {
		return &schemaTypeList{OneOf: types}
	}
}

type schemaReference struct {
	schemaTypeWithoutBase
	Ref string `json:"$ref"`
}

func newSchemaBool() *schemaTypeBase {
	return withBaseType(new(schemaTypeBase), "boolean")
}

type schemaObject struct {
	schemaTypeBase
	AdditionalProperties any                   `json:"additionalProperties,omitempty"`
	PatternProperties    map[string]SchemaType `json:"patternProperties,omitempty"`
	Properties           map[string]SchemaType `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	DependentRequired    map[string][]string   `json:"dependentRequired,omitempty"`
}

func newSchemaObject() *schemaObject {
	return withBaseType(&schemaObject{
		AdditionalProperties: false,
		PatternProperties:    make(map[string]SchemaType),
		Properties:           make(map[string]SchemaType),
		DependentRequired:    make(map[string][]string),
	}, "object")
}

func (s *schemaObject) verify() (err error) {
	for pattern, st := range s.PatternProperties {
		if err != nil {
			break
		} else if st == nil {
			err = fmt.Errorf("type of %q in patternProperties is undefined", pattern)
		} else if err = verifySchemaRegex(pattern); err != nil {
			err = fmt.Errorf("patternProperties regex %q failed to compile: %w", pattern, err)
		}
	}
	for name, st := range s.Properties {
		if err != nil {
			break
		} else if st == nil {
			err = fmt.Errorf("type of %q in properties is undefined", name)
		}
	}
	if err == nil {
		switch s.AdditionalProperties.(type) {
		case nil:
		case bool:
		case SchemaType:
		default:
			err = fmt.Errorf("additionalProperties must be nil, boolean or SchemaType")
		}
	}
	if err == nil {
		err = s.schemaTypeBase.verify()
	}
	return
}

type schemaArray struct {
	schemaTypeBase
	Items       SchemaType `json:"items"`
	MinItems    uint64     `json:"minItems"`
	MaxItems    *uint64    `json:"maxItems,omitempty"`
	UniqueItems bool       `json:"uniqueItems"`
}

func newSchemaArray(items SchemaType) *schemaArray {
	return withBaseType(&schemaArray{Items: items}, "array")
}

func (a *schemaArray) describe(title, description string) {
	a.schemaTypeBase.describe(title, description)
	a.Title = title + arrayTitleSuffix
}

func (a *schemaArray) verify() (err error) {
	if a.Items == nil {
		err = fmt.Errorf("items of schemaArray is undefined")
	} else {
		err = a.schemaTypeBase.verify()
	}
	return
}

type schemaNumber struct {
	schemaTypeBase
	MultipleOf       *float64 `json:"multipleOf,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
}

func newSchemaNumber(integer bool) *schemaNumber {
	typeName := "number"
	if integer {
		typeName = "integer"
	}
	return withBaseType(new(schemaNumber), typeName)
}

type stringFormat string

const (
	FmtAny      = stringFormat("")
	FmtDateTime = stringFormat("date-time")
	FmtDate     = stringFormat("date")
	FmtTime     = stringFormat("time")
	FmtEmail    = stringFormat("email")
	FmtHostname = stringFormat("hostname")
	FmtIPv4     = stringFormat("ipv4")
	FmtIPv6     = stringFormat("ipv6")
	FmtURI      = stringFormat("uri")
	FmtURIRef   = stringFormat("uri-reference")
	FmtUUID     = stringFormat("uuid")
	FmtRegex    = stringFormat("regex")
)

var validFormatNames = []stringFormat{
	FmtAny, FmtDateTime, FmtDate, FmtTime, FmtEmail,
	FmtHostname, FmtIPv4, FmtIPv6, FmtURI, FmtURIRef, FmtUUID, FmtRegex,
}

type schemaString struct {
	schemaTypeBase
	MinLength        uint64       `json:"minLength"`
	MaxLength        *uint64      `json:"maxLength,omitempty"`
	ContentEncoding  string       `json:"contentEncoding,omitempty"`
	ContentMediaType string       `json:"contentMediaType,omitempty"`
	Pattern          string       `json:"pattern,omitempty"`
	Format           stringFormat `json:"format,omitempty"`
}

func (s *schemaString) verify() (err error) {
	if len(s.Pattern) > 0 {
		if err = verifySchemaRegex(s.Pattern); err != nil {
			err = fmt.Errorf("pattern regex %q failed to compile: %w", s.Pattern, err)
		}
	}

	if !slices.Contains(validFormatNames, s.Format) {
		err = fmt.Errorf("format %q is no valid string format", s.Format)
	}
	return
}

func newSchemaString() *schemaString {
	return withBaseType(new(schemaString), "string")
}

func newSchemaBase64String(mediaType string) *schemaString {
	s := newSchemaString()
	s.ContentEncoding = "base64"
	s.ContentMediaType = mediaType
	return s
}

func verifySchemaRegex(pattern string) (err error) {
	pattern = strings.ReplaceAll(pattern, "(?!", "(")
	_, err = regexp.Compile(pattern)
	return
}

// walkTypes walks into all SchemaType items from start, passing them to callback (including start).
// The callback can return item, nil or a new SchemaType as it would like to continue to walk deeper,
// stop walking deeper or replace item with a new SchemaType.
func walkTypes(start SchemaType, callback func(item SchemaType) SchemaType) SchemaType {
	return internalWalkTypes(make(map[SchemaType]bool), start, callback)
}

func internalWalkTypes(into map[SchemaType]bool, current SchemaType, callback func(item SchemaType) SchemaType) SchemaType {
	if !into[current] && current != nil {
		defer delete(into, current)
		into[current] = true

		if processed := callback(current); processed == nil { // callback tells we must not walk deeper
			return current
		} else if current != processed { // callback replaced current
			defer delete(into, processed)
			into[processed] = true
			current = processed
		}

		switch t := current.(type) {
		case nil:
			// ignore
		case *schemaObject:
			for name, property := range t.Properties {
				t.Properties[name] = internalWalkTypes(into, property, callback)
			}
			for name, property := range t.PatternProperties {
				t.PatternProperties[name] = internalWalkTypes(into, property, callback)
			}
			if item, ok := t.AdditionalProperties.(SchemaType); ok {
				t.AdditionalProperties = internalWalkTypes(into, item, callback)
			}
		case *schemaArray:
			t.Items = internalWalkTypes(into, t.Items, callback)
		case *schemaTypeList:
			for index, property := range t.OneOf {
				t.OneOf[index] = internalWalkTypes(into, property, callback)
			}
			for index, property := range t.AnyOf {
				t.AnyOf[index] = internalWalkTypes(into, property, callback)
			}
		}
	}

	return current
}
