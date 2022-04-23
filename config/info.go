package config

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/util"
	"golang.org/x/exp/maps"
)

// ProfileInfo provides structural information on profiles and can be used for specification and validation
//
//go:generate mockery --name=ProfileInfo
type ProfileInfo interface {
	// PropertySet contains properties shared across sections
	PropertySet
	// Sections returns a list of all known section names
	Sections() []string
	// SectionInfo returns information for a named section
	SectionInfo(name string) SectionInfo
}

// SectionInfo provides structural information on a particular profile section
//
//go:generate mockery --name=SectionInfo
type SectionInfo interface {
	// Named provides Name and Description of the section
	Named
	// PropertySet contains properties for this section
	PropertySet
	// IsCommandSection indicates whether the section configures a restic command
	IsCommandSection() bool
	// Command provides the restic command for the section when this is a command section
	Command() restic.CommandIf
}

// Named provides Name and Description for an item
type Named interface {
	// Name provides the name of the item
	Name() string
	// Description provides a possibly empty description of the item
	Description() string
}

// PropertySet provides structural information on a set of properties
type PropertySet interface {
	// TypeName returns the name of the type that declared the property set. The name may be empty.
	TypeName() string
	// IsClosed indicates that a property set supports only know properties (as opposed to a non-closed set that may hold arbitrary properties)
	IsClosed() bool
	// IsAllOptions indicates that all properties in this set return IsOption true
	IsAllOptions() bool
	// Properties provides the names of all known config properties in the set
	Properties() []string
	// PropertyInfo provides information on a named property
	PropertyInfo(name string) PropertyInfo
	// OtherPropertyInfo provides information on arbitrary properties in a set where IsClosed is false. Is nil if any arbitrary property is supported.
	OtherPropertyInfo() PropertyInfo
}

// NamedPropertySet is Named and PropertySet
//
//go:generate mockery --name=NamedPropertySet
type NamedPropertySet interface {
	Named
	PropertySet
}

// PropertyInfo provides details on an individual property
//
//go:generate mockery --name=PropertyInfo
type PropertyInfo interface {
	// Named provides Name and Description of the property
	Named
	// IsOption indicates whether the property is a restic command option
	IsOption() bool
	// Option provides the restic option when IsOption is true
	Option() restic.Option
	// IsRequired indicates whether the property must be present in its PropertySet
	IsRequired() bool
	// IsDeprecated indicates whether the property is discontinued
	IsDeprecated() bool
	// IsSingle indicates that the property can be defined only once.
	IsSingle() bool
	// IsMultiType indicates that more than one of CanBeString, CanBeNumeric, CanBeBool & CanBePropertySet returns true
	IsMultiType() bool
	// IsAnyType indicates that all of CanBeString, CanBeNumeric & CanBeBool return true
	IsAnyType() bool
	// CanBeString indicates that the value can a string value
	CanBeString() bool
	// CanBeNumeric indicates that the value can be numeric
	CanBeNumeric() bool
	// CanBeBool indicates that the value can be boolean true or false
	CanBeBool() bool
	// CanBeNil indicates that the value can be set to null (or undefined)
	CanBeNil() bool
	// CanBePropertySet indicates that the property can be a property set (= nested object or array of objects)
	CanBePropertySet() bool
	// PropertySet returns the property set that this property can be filled with (possibly nil)
	PropertySet() NamedPropertySet
	// Format may provide a string describing the expected input format (e.g. time, duration, base64, etc.).
	Format() string
	// ValidationPattern may provide a regular expression to validate the value
	ValidationPattern() string
	// MustBeInteger indicates that a numeric value must be integer
	MustBeInteger() bool
	// DefaultValue returns the default value(s)
	DefaultValue() []string
	// ExampleValues may provide example values
	ExampleValues() []string
	// EnumValues may provide a list of all possible values
	EnumValues() []string
	// NumericRange may provide a valid range for numbers
	NumericRange() NumericRange
}

// NumericRange holds a numeric range constraint of a PropertyInfo
type NumericRange struct {
	From, To                   *float64
	FromExclusive, ToExclusive bool
}

// profileInfo implements ProfileInfo
type profileInfo struct {
	propertySet
	sections map[string]SectionInfo
}

func (p *profileInfo) SectionInfo(name string) SectionInfo { return p.sections[name] }

func (p *profileInfo) Sections() (names []string) {
	if names = maps.Keys(p.sections); names != nil {
		sort.Strings(names)
	}
	return
}

// sectionInfo implements SectionInfo
type sectionInfo struct {
	namedPropertySet
	command restic.CommandIf
}

func (s *sectionInfo) IsCommandSection() bool    { return s.command != nil }
func (s *sectionInfo) Command() restic.CommandIf { return s.command }

// propertySet implements PropertySet
type propertySet struct {
	openSet       bool
	typeName      string
	properties    map[string]PropertyInfo
	otherProperty PropertyInfo
}

func (p *propertySet) TypeName() string                      { return p.typeName }
func (p *propertySet) IsClosed() bool                        { return !p.openSet }
func (p *propertySet) PropertyInfo(name string) PropertyInfo { return p.properties[name] }
func (p *propertySet) OtherPropertyInfo() PropertyInfo       { return p.otherProperty }

func (p *propertySet) Properties() (names []string) {
	if names = maps.Keys(p.properties); names != nil {
		sort.Strings(names)
	}
	return
}

func (p *propertySet) IsAllOptions() bool {
	for _, info := range p.properties {
		if !info.IsOption() {
			return false
		}
	}
	return true
}

// namedPropertySet extends propertySet with Named
type namedPropertySet struct {
	propertySet
	name, description string
}

func (p *namedPropertySet) Name() string        { return p.name }
func (p *namedPropertySet) Description() string { return p.description }

// accessibleProperty provides package local access to basicPropertyInfo and the backing struct field (when available)
type accessibleProperty interface {
	// sectionField returns or sets the field that declares the PropertySet that this property is in, or nil if unknown.
	sectionField(*reflect.StructField) *reflect.StructField
	// field returns the field that declares this property, or nil if the property is not based on a field
	field() *reflect.StructField
	// basic returns the mutable basicPropertyInfo (is never nil)
	basic() *basicPropertyInfo
	// option returns the underlying restic.Option or nil if the property is not of type resticPropertyInfo.
	option() *restic.Option
}

// basicPropertyInfo is the base for PropertyInfo implementations
type basicPropertyInfo struct {
	mayString, mayNumber, mayBool, mayNil, mustInt bool
	deprecated, required, single                   bool
	from, to                                       *float64
	fromExclusive, toExclusive                     bool
	name, format, pattern                          string
	examples, enum                                 []string
	nested                                         NamedPropertySet
	descriptionFilter                              func(string) string
}

func (b *basicPropertyInfo) Name() string                  { return b.name }
func (b *basicPropertyInfo) IsDeprecated() bool            { return b.deprecated }
func (b *basicPropertyInfo) IsRequired() bool              { return b.required }
func (b *basicPropertyInfo) IsSingle() bool                { return b.single }
func (b *basicPropertyInfo) CanBeBool() bool               { return b.mayBool }
func (b *basicPropertyInfo) CanBeNil() bool                { return b.mayNil }
func (b *basicPropertyInfo) CanBeNumeric() bool            { return b.mayNumber }
func (b *basicPropertyInfo) CanBeString() bool             { return b.mayString }
func (b *basicPropertyInfo) CanBePropertySet() bool        { return b.nested != nil }
func (b *basicPropertyInfo) PropertySet() NamedPropertySet { return b.nested }
func (b *basicPropertyInfo) MustBeInteger() bool           { return b.mustInt }
func (b *basicPropertyInfo) Format() string                { return b.format }
func (b *basicPropertyInfo) ValidationPattern() string     { return b.pattern }
func (b *basicPropertyInfo) ExampleValues() []string       { return append([]string{}, b.examples...) }
func (b *basicPropertyInfo) EnumValues() []string          { return append([]string{}, b.enum...) }
func (b *basicPropertyInfo) field() *reflect.StructField   { return nil }
func (b *basicPropertyInfo) basic() *basicPropertyInfo     { return b }
func (b *basicPropertyInfo) option() *restic.Option        { return nil }

func (b *basicPropertyInfo) filterDescription(description string) string {
	if b.descriptionFilter != nil {
		return b.descriptionFilter(description)
	}
	return description
}

func (b *basicPropertyInfo) addDescriptionFilter(fn func(description string) string) {
	if b.descriptionFilter == nil {
		b.descriptionFilter = fn
	} else if fn != nil {
		previous := b.descriptionFilter
		b.descriptionFilter = func(desc string) string { return fn(previous(desc)) }
	}
}

func (b *basicPropertyInfo) NumericRange() NumericRange {
	return NumericRange{
		From:          b.from,
		FromExclusive: b.from == nil || b.fromExclusive,
		To:            b.to,
		ToExclusive:   b.to == nil || b.toExclusive,
	}
}

func (b *basicPropertyInfo) sectionField(f *reflect.StructField) *reflect.StructField { return f }

func (b *basicPropertyInfo) IsMultiType() bool {
	return b.countTrue(b.CanBeBool(), b.CanBeNumeric(), b.CanBeString(), b.CanBePropertySet()) > 1
}

func (b *basicPropertyInfo) IsAnyType() bool {
	return b.countTrue(b.CanBeBool(), b.CanBeNumeric(), b.CanBeString()) == 3
}

func (b *basicPropertyInfo) countTrue(args ...bool) (c int) {
	for _, tr := range args {
		if tr {
			c++
		}
	}
	return
}

func (b *basicPropertyInfo) resetTypeInfo() *basicPropertyInfo {
	b.mayBool = false
	b.mayNumber = false
	b.mustInt = false
	b.mayString = false
	b.nested = nil
	return b
}

// propertyInfo implements PropertyInfo
type propertyInfo struct {
	basicPropertyInfo
	description    string
	defaults       []string
	originField    *reflect.StructField
	fieldInSection *reflect.StructField
}

func (p *propertyInfo) Description() string           { return p.filterDescription(p.description) }
func (p *propertyInfo) DefaultValue() []string        { return append([]string{}, p.defaults...) }
func (p *propertyInfo) IsOption() bool                { return false }
func (p *propertyInfo) Option() (empty restic.Option) { return }
func (p *propertyInfo) field() *reflect.StructField   { return p.originField }
func (p *propertyInfo) sectionField(f *reflect.StructField) *reflect.StructField {
	if f != nil {
		p.fieldInSection = f
	}
	if p.fieldInSection == nil {
		return p.field()
	}
	return p.fieldInSection
}

// resticPropertyInfo implements PropertyInfo for properties derived from restic.Option
type resticPropertyInfo struct {
	basicPropertyInfo
	originField      *reflect.StructField
	optionFlag       restic.Option
	skipDefaultValue bool
}

func (r *resticPropertyInfo) DefaultValue() (defaults []string) {
	if len(r.optionFlag.Default) > 0 && !r.skipDefaultValue {
		defaults = append(defaults, r.optionFlag.Default)
	}
	return
}

func (r *resticPropertyInfo) Description() string {
	return r.filterDescription(r.optionFlag.Description)
}

func (r *resticPropertyInfo) IsOption() bool              { return true }
func (r *resticPropertyInfo) Option() restic.Option       { return r.optionFlag }
func (r *resticPropertyInfo) option() *restic.Option      { return &r.optionFlag }
func (p *resticPropertyInfo) field() *reflect.StructField { return p.originField }

var (
	numberValuePattern = regexp.MustCompile(`^-?\d?\.?\d+$`)
	intValuePattern    = regexp.MustCompile(`^-?\d+$`)
	boolValuePattern   = regexp.MustCompile(`^(true|false)$`)
	formatPattern      = regexp.MustCompile(`^(date-time|date|time|duration|email|hostname|ipv4|ipv6|uri|uri-reference|uuid|regex)$`)
	rangePattern       = regexp.MustCompile(`^([\[|])\s*([\d.\-]*)\s*:\s*([\d.\-]*)\s*([|\]])$`)
)

func newResticPropertyInfo(name string, option restic.Option) *resticPropertyInfo {
	info := &resticPropertyInfo{
		basicPropertyInfo: basicPropertyInfo{
			name:       name,
			required:   false,
			single:     option.Once,
			deprecated: len(option.RemovedInVersion) > 0,
			mustInt:    intValuePattern.MatchString(option.Default),
		},
		optionFlag: option,
	}

	if info.MustBeInteger() || numberValuePattern.MatchString(option.Default) {
		info.mayNumber = true
	} else if boolValuePattern.MatchString(option.Default) {
		info.mayBool = true
	} else {
		info.mayString = true
	}

	return info
}

func propertySetFromType(t reflect.Type) (props propertySet) {
	props.typeName = t.Name()
	props.properties = make(map[string]PropertyInfo)
	for i, num := 0, t.NumField(); i < num; i++ {
		field := t.Field(i)

		if ms, found := field.Tag.Lookup("mapstructure"); found {
			if strings.Contains(ms, ",squash") {
				inlineSet := propertySetFromType(field.Type)
				for name, p := range inlineSet.properties {
					props.properties[name] = p
					if ap, ok := p.(accessibleProperty); ok {
						ap.sectionField(&field)
					}
				}
				if inlineSet.openSet {
					props.openSet = true
					props.otherProperty = inlineSet.otherProperty
				}
			} else {
				p := propertyInfo{originField: &field}
				configurePropertyFromType(p.basic(), field.Type)
				configurePropertyFromTags(&p, &field)

				if strings.Contains(ms, ",remain") {
					props.openSet = true
					if p.CanBePropertySet() {
						props.otherProperty = p.PropertySet().OtherPropertyInfo()
					}
				} else {
					props.properties[p.Name()] = &p
				}
			}
		}
	}
	return
}

func configurePropertyFromType(p *basicPropertyInfo, valueType reflect.Type) {
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
		p.mayNil = true
	}

	if valueType.Kind() == reflect.Array || valueType.Kind() == reflect.Slice {
		valueType = valueType.Elem()
	} else {
		p.single = true
	}

	p.mayString = true

	switch valueType.Kind() {
	case reflect.Bool:
		p.resetTypeInfo().mayBool = true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p.resetTypeInfo().mayNumber = true
		p.mustInt = true
		if valueType == reflect.TypeOf(time.Duration(0)) {
			p.format = "duration"
			p.mayString = true
		}
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		p.resetTypeInfo().mayNumber = true
	case reflect.Struct:
		p.resetTypeInfo().nested = &namedPropertySet{
			propertySet: propertySetFromType(valueType),
			name:        p.Name(),
		}
	case reflect.Map:
		set := &namedPropertySet{
			propertySet: propertySet{openSet: true},
			name:        p.Name(),
		}
		setValues := &propertyInfo{}
		configurePropertyFromType(setValues.basic(), util.ElementType(valueType))
		if !setValues.IsAnyType() {
			set.otherProperty = setValues
		}
		p.resetTypeInfo().nested = set
		p.mayNil = true
	case reflect.Interface:
		p.mayString = true
		p.mayNumber = true
		p.mayBool = true
		p.mayNil = true
	}
}

// joinEmpty joins strings tokens around an empty token, e.g. ["a", "start", "", "end", "b"] turns to ["a", "start-end", "b"] (glue -)
// can be used to split strings by a delimiter and allow delimiter escaping by repeating the delimiter
func joinEmpty(input []string, glue string) (output []string) {
	add := false
	output = make([]string, 0, len(input))
	for i, s := range input {
		if s == "" && i > 0 && i < len(input)-1 {
			if add {
				output = append(output, s)
			} else {
				output[len(output)-1] += glue
				add = true
			}
		} else if add {
			output[len(output)-1] += s
			add = false
		} else {
			output = append(output, s)
		}
	}
	return
}

func configurePropertyFromTags(p *propertyInfo, field *reflect.StructField) {
	// description tag e.g. `description:"the quiet setting disables all output from resticprofile"`
	if value, found := field.Tag.Lookup("description"); found {
		p.description = value
	}
	// default value tag e.g. `default:"DefaultValue"` or `default:"Item1;Item2;Item3"`
	if value, found := field.Tag.Lookup("default"); found && len(value) > 0 {
		p.defaults = joinEmpty(strings.Split(value, ";"), ";")
	} else if !found && !p.IsMultiType() && !p.CanBeNil() {
		if p.CanBeBool() {
			p.defaults = []string{"false"}
		} else if p.CanBeNumeric() {
			p.defaults = []string{"0"}
		}
	}
	configureBasicPropertyFromTags(&p.basicPropertyInfo, field)
}

func configureBasicPropertyFromTags(p *basicPropertyInfo, field *reflect.StructField) {
	if value, found := field.Tag.Lookup("mapstructure"); found {
		p.name = strings.Split(value, ",")[0]
	}
	// deprecated marker
	if _, found := field.Tag.Lookup("deprecated"); found {
		p.deprecated = true
	}
	// example value tag e.g. `examples:"ExampleValue"` or `examples:"EV1;EV2;EV3"`
	if value, found := field.Tag.Lookup("examples"); found && len(value) > 0 {
		p.examples = joinEmpty(strings.Split(value, ";"), ";")
	}
	// enum value tag e.g. `enum:"PossibleValue1;PossibleValue2;PossibleValue3"`
	if value, found := field.Tag.Lookup("enum"); found && len(value) > 0 {
		p.enum = joinEmpty(strings.Split(value, ";"), ";")
	}
	// format tag e.g. `format:"time"`
	if value, found := field.Tag.Lookup("format"); found && formatPattern.MatchString(value) {
		p.format = value
	}
	// value validation tag e.g. `pattern:".*Regex.*"`
	if value, found := field.Tag.Lookup("pattern"); found {
		p.pattern = value
	}
	// value validation tag e.g. `range:"[-6:10]"` (min -6 to max 10) or `range:"[:10|" (min inf - max 10 exclusive)`
	if value, found := field.Tag.Lookup("range"); found && rangePattern.MatchString(value) {
		if m := rangePattern.FindStringSubmatch(value); m != nil && len(m) == 5 {
			if from, err := strconv.ParseFloat(m[2], 64); len(m[2]) > 0 && err == nil {
				p.from = &from
			}
			if to, err := strconv.ParseFloat(m[3], 64); len(m[3]) > 0 && err == nil {
				p.to = &to
			}
			p.fromExclusive = m[1] == "|"
			p.toExclusive = m[4] == "|"
		}
	}
}

// propertyCustomizer allows to adjust properties before they get appended to a section
type propertyCustomizer func(sectionName, propertyName string, property accessibleProperty)

var (
	propertyCustomizers []propertyCustomizer
	exclusions          []string
)

// registerPropertyInfoCustomizer registers a callback that may modify properties
func registerPropertyInfoCustomizer(c propertyCustomizer) {
	propertyCustomizers = append(propertyCustomizers, c)
}

// ExcludeProfileSection allows to exclude a section from the profile.
// To be used for overlapping commands used by restic and resticprofile.
func ExcludeProfileSection(sectionName string) {
	ExcludeProfileProperty(sectionName, "*")
}

// ExcludeProfileProperty allows to exclude a section from the profile.
// To be used for overlapping commands used by restic and resticprofile.
func ExcludeProfileProperty(sectionName, propertyName string) {
	exclusions = append(exclusions, fmt.Sprintf("%s.%s", sectionName, propertyName))
	sort.Strings(exclusions)
}

func isExcluded(sectionName, propertyName string) bool {
	exclusion := fmt.Sprintf("%s.%s", sectionName, propertyName)
	index := sort.SearchStrings(exclusions, exclusion)
	return index < len(exclusions) && exclusions[index] == exclusion
}

// customizeProperties customizes a PropertySet and applies propertyCustomizers and exclusions
func customizeProperties(sectionName string, properties map[string]PropertyInfo) {
	for propertyName, property := range properties {
		if isExcluded(sectionName, propertyName) || isExcluded("*", propertyName) || property == nil {
			delete(properties, propertyName)
		} else {
			for _, customizer := range propertyCustomizers {
				if ap, ok := property.(accessibleProperty); ok {
					customizer(sectionName, propertyName, ap)
				}
			}
		}
	}
}

var infoTypes struct {
	global,
	group,
	mixins,
	mixinUse,
	profile,
	genericSection reflect.Type
	genericSectionNames []string
}

func init() {
	// Init infoTypes
	{
		profile := *NewProfile(nil, "")
		infoTypes.global = reflect.TypeOf(Global{})
		infoTypes.group = reflect.TypeOf(Group{})
		infoTypes.mixins = reflect.TypeOf(mixin{})
		infoTypes.mixinUse = reflect.TypeOf(mixinUse{})
		infoTypes.profile = reflect.TypeOf(profile)
		infoTypes.genericSection = reflect.TypeOf(GenericSection{})
		infoTypes.genericSectionNames = maps.Keys(profile.OtherSections)
	}
}

// NewGlobalInfo returns structural information on the "global" config section
func NewGlobalInfo() NamedPropertySet {
	set := &namedPropertySet{
		name:        constants.SectionConfigurationGlobal,
		description: "global settings",
		propertySet: propertySetFromType(infoTypes.global),
	}
	customizeProperties(constants.SectionConfigurationGlobal, set.properties)
	return set
}

// NewGroupInfo returns structural information on the "group" config v2 section
func NewGroupInfo() NamedPropertySet {
	return &namedPropertySet{
		name:        constants.SectionConfigurationGroups,
		description: "profile groups",
		propertySet: propertySetFromType(infoTypes.group),
	}
}

// NewMixinsInfo returns structural information on the "mixins" config v2 section
func NewMixinsInfo() NamedPropertySet {
	return &namedPropertySet{
		name:        constants.SectionConfigurationMixins,
		description: "global mixins declaration",
		propertySet: propertySetFromType(infoTypes.mixins),
	}
}

// NewMixinUseInfo returns structural information on the mixin "use" flags in config v2
func NewMixinUseInfo() NamedPropertySet {
	return &namedPropertySet{
		name:        constants.SectionConfigurationMixinUse,
		description: "named mixin reference to apply to the current location",
		propertySet: propertySetFromType(infoTypes.mixinUse),
	}
}

// NewProfileInfo returns structural information on the "profile" config section
func NewProfileInfo(withDefaultOptions bool) ProfileInfo {
	return NewProfileInfoForRestic(restic.AnyVersion, withDefaultOptions)
}

// NewProfileInfoForRestic returns versioned structural information on the "profile" config section
// for the specified semantic resticVersion. withDefaultOptions toggles whether to include default
// command options with every single section.
func NewProfileInfoForRestic(resticVersion string, withDefaultOptions bool) ProfileInfo {
	// Building initial set including generic sections (from data model)
	profileSet := propertySetFromType(infoTypes.profile)
	{
		genericSection := propertySetFromType(infoTypes.genericSection)
		for _, name := range infoTypes.genericSectionNames {
			pi := new(propertyInfo)
			pi.nested = &namedPropertySet{
				propertySet: genericSection,
				name:        name,
			}
			profileSet.properties[name] = pi
		}
	}

	// Building sections (from restic info)
	nameAliases := make(map[string]string)
	sections := make(map[string]SectionInfo)
	defaultCommand, _ := restic.GetCommandForVersion(restic.DefaultCommand, resticVersion, false)

	addSection := func(sectionName string, command restic.CommandIf, properties map[string]PropertyInfo) {
		// remove initial property set
		delete(profileSet.properties, sectionName)

		// customization / exclusion
		customizeProperties(sectionName, properties)

		// add section
		if len(properties) > 0 && !isExcluded(sectionName, "*") {
			section := &sectionInfo{
				command: command,
				namedPropertySet: namedPropertySet{
					propertySet: propertySet{properties: properties},
					name:        sectionName,
					description: command.GetDescription(),
				},
			}

			sections[sectionName] = section
		}
	}

	customizeCommandOption := func(info PropertyInfo, command restic.CommandIf) PropertyInfo {
		if ap, ok := info.(accessibleProperty); ok && ap.field() != nil {
			if optionName, ok := ap.field().Tag.Lookup("argument"); ok {
				optionName = strings.Split(optionName, ",")[0]
				if option, found := command.Lookup(optionName); found {
					nameAliases[option.Name] = info.Name()
					p := newResticPropertyInfo(info.Name(), option)
					if p.originField = ap.field(); p.originField != nil {
						configurePropertyFromType(&p.basicPropertyInfo, p.originField.Type)
						configureBasicPropertyFromTags(&p.basicPropertyInfo, p.originField)
					}
					info = p
				} else {
					return nil // unknown restic flag, remove it
				}
			}
		}
		return info
	}

	addMissingCommandOptions := func(command restic.CommandIf, skipDefaultValue bool, properties map[string]PropertyInfo) {
		for _, option := range command.GetOptions() {
			name := option.Name
			if alias, ok := nameAliases[name]; ok {
				name = alias
			}
			if _, found := properties[name]; !found {
				info := newResticPropertyInfo(name, option)
				info.skipDefaultValue = skipDefaultValue
				properties[name] = info
			}
		}
	}

	// Map struct fields against restic commands and options
	for name, property := range profileSet.properties {
		if !property.CanBePropertySet() || property.IsMultiType() {
			continue
		}

		commandName := name
		if ap, ok := property.(accessibleProperty); ok && ap.field() != nil {
			if c, ok := ap.field().Tag.Lookup("command"); ok {
				commandName = c
			}
		}

		// Translate command sections from the profile's PropertySet
		if command, found := restic.GetCommandForVersion(commandName, restic.AnyVersion, true); found {
			set := property.PropertySet()
			sectionProperties := make(map[string]PropertyInfo)

			// Populate sectionProperties if restic hash some at the selected target version
			if command, found = restic.GetCommandForVersion(command.GetName(), resticVersion, false); found {
				// Check properties from the struct
				for _, nestedName := range set.Properties() {
					info := set.PropertyInfo(nestedName)
					if info = customizeCommandOption(info, command); info != nil {
						sectionProperties[nestedName] = info
					}
				}

				// Add dynamic options not defined in the struct
				addMissingCommandOptions(command, false, sectionProperties)
				if withDefaultOptions && defaultCommand != nil {
					// skipping default value since the default is in the profile (not restic's default)
					addMissingCommandOptions(defaultCommand, true, sectionProperties)
				}
			}

			// Add section (or remove it when sectionProperties are left empty)
			addSection(name, command, sectionProperties)
		}
	}

	// Customize profile properties after sections have been built (must not run earlier, naming conflicts might arise)
	if defaultCommand != nil {
		for name, info := range profileSet.properties {
			if info = customizeCommandOption(info, defaultCommand); info != nil {
				profileSet.properties[name] = info
			}
		}
		// Add dynamic options not defined in the struct
		addMissingCommandOptions(defaultCommand, false, profileSet.properties)

		customizeProperties("", profileSet.properties)
	}

	return &profileInfo{propertySet: profileSet, sections: sections}
}
