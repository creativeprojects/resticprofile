package jsonschema

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/spf13/cast"
)

// WriteJsonSchema generates a JSON schema to validate (and complete) resticprofile JSON and YAML configuration files.
// Use version and resticVersion to specify what config file format and restic version to use when generating the schema.
func WriteJsonSchema(version config.Version, resticVersion string, output io.Writer) error {
	if schema, err := generateJsonSchema(version, resticVersion); err == nil {
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(true)
		return encoder.Encode(schema)
	} else {
		return fmt.Errorf("no JSON schema was created: %w", err)
	}
}

const matchAll = ".+"

var typePatterns struct {
	boolean,
	integer,
	number,
	string *regexp.Regexp
}

func init() {
	typePatterns.boolean = regexp.MustCompile(`(?i)^(true|false)$`)
	typePatterns.integer = regexp.MustCompile(`^-?\d+$`)
	typePatterns.number = regexp.MustCompile(`^-?[.,\d^]+$`)
	typePatterns.string = regexp.MustCompile(`^".*"$`)
}

func convertToType(input string) any {
	if typePatterns.boolean.MatchString(input) {
		return cast.ToBool(input)
	} else if typePatterns.integer.MatchString(input) {
		return cast.ToInt64(input)
	} else if typePatterns.number.MatchString(input) {
		return cast.ToFloat64(input)
	} else if typePatterns.string.MatchString(input) {
		return input[1 : len(input)-1]
	}
	return input
}

func isCompatibleValue(schema SchemaType, value any) (ok bool) {
	if base := schema.base(); base != nil {
		switch base.Type {
		case "boolean":
			_, ok = value.(bool)
		case "integer":
			_, ok = value.(int64)
		case "number":
			_, ok = value.(int64)
			if !ok {
				_, ok = value.(float64)
			}
		default:
			ok = true
		}
	}
	return
}

func isDefaultValueForType(value any) bool {
	switch v := value.(type) {
	case bool:
		return v == false
	case int64:
		return v == 0
	case float64:
		return v == 0
	case string:
		return len(v) == 0
	}
	return false
}

func getDescription(info config.PropertyInfo) (description string) {
	description = info.Description()

	var notes []string
	if info.IsDeprecated() {
		notes = append(notes, "deprecated")
	}

	if info.IsOption() {
		option := info.Option()
		if len(option.RemovedInVersion) > 0 {
			notes = append(notes, fmt.Sprintf("removed in %s", option.RemovedInVersion))
		}
		if len(option.FromVersion) > 0 {
			notes = append(notes, fmt.Sprintf("restic >= %s", option.FromVersion))
		}
	}

	if len(notes) > 0 {
		description = fmt.Sprintf("%s [%s]", description, strings.Join(notes, ", "))
	}
	return
}

func schemaForPropertySet(props config.PropertySet) (object *schemaObject) {
	object = newSchemaObject()
	object.AdditionalProperties = !props.IsClosed()
	if np, ok := props.(config.Named); ok {
		object.describe(np.Name(), np.Description())
	}

	if !props.IsClosed() {
		object.AdditionalProperties = true

		if info := props.OtherPropertyInfo(); info != nil && len(props.Properties()) == 0 {
			if propertyType := schemaForPropertyInfo(info); propertyType != nil {
				object.PatternProperties[matchAll] = propertyType
				object.AdditionalProperties = false
			}
		}
	}

	for _, propertyName := range props.Properties() {
		info := props.PropertyInfo(propertyName)
		propertyType := schemaForPropertyInfo(info)
		if propertyType == nil {
			continue
		}

		// Append to object
		object.Properties[propertyName] = propertyType

		// Check if this property is required
		if info.IsRequired() {
			object.Required = append(object.Required, propertyName)
		}
	}

	return
}

func schemaForPropertyInfo(info config.PropertyInfo) SchemaType {
	// Detect item type of this property
	var propertyType, nestedType SchemaType

	if types, nestedIndex := typesFromPropertyInfo(info); len(types) > 0 {
		if len(types) > 1 {
			oneOf := info.IsSingle()
			propertyType = newSchemaTypeList(!oneOf, types...)
		} else {
			propertyType = types[0]
		}
		if nestedIndex > -1 {
			nestedType = types[nestedIndex]
		}
	} else {
		return nil
	}

	// Array or single type
	if !info.IsSingle() {
		// viper supports single elements for list types
		propertyType = newSchemaTypeList(true, propertyType, newSchemaArray(propertyType))
	}

	// Set basic info
	configureBasicInfo(propertyType, nestedType, info)

	return propertyType
}

// durationPattern is a regex pattern that validates the go duration string format
const durationPattern = `^(\d+(h|m|s|ms))+$`

func typesFromPropertyInfo(info config.PropertyInfo) (types []SchemaType, nestedTypeIndex int) {
	if info == nil {
		return
	}
	if info.CanBeBool() {
		types = append(types, newSchemaBool())
	}
	if info.CanBeNumeric() {
		number := newSchemaNumber(info.MustBeInteger())
		nr := info.NumericRange()
		if nr.From != nil {
			if nr.FromExclusive {
				number.ExclusiveMinimum = nr.From
			} else {
				number.Minimum = nr.From
			}
		}
		if nr.To != nil {
			if nr.ToExclusive {
				number.ExclusiveMaximum = nr.To
			} else {
				number.Maximum = nr.To
			}
		}
		types = append(types, number)
	}
	if info.CanBeString() {
		s := newSchemaString()
		if info.Format() == "duration" {
			// using custom validation for duration as JSONSchema's duration is not compatible with go
			s.Pattern = durationPattern
			s.MinLength = 2
		} else {
			s.Format = stringFormat(info.Format())
		}
		if pattern := info.ValidationPattern(); len(pattern) > 0 {
			s.Pattern = pattern
		}
		types = append(types, s)
	}
	if info.CanBePropertySet() {
		nestedTypeIndex = len(types)
		types = append(types, schemaForPropertySet(info.PropertySet()))
	} else {
		nestedTypeIndex = -1
	}
	return
}

func configureBasicInfo(propertyType, nestedType SchemaType, info config.PropertyInfo) {
	defaultValue := collect.From(info.DefaultValue(), convertToType)
	exampleValues := collect.From(info.ExampleValues(), convertToType)
	enumValues := collect.From(info.EnumValues(), convertToType)
	description := getDescription(info)

	walkTypes(propertyType, func(item SchemaType) SchemaType {
		if item == nestedType {
			return nil // do not walk into the nested type (if defined)
		}

		item.describe(info.Name(), description)

		if base := item.base(); base != nil {
			base.setDeprecated(info.IsDeprecated())

			// Skip defaults for deprecated values (to avoid they are pre-filled in auto-complete)
			if !info.IsDeprecated() {
				defaults := collect.All(defaultValue, func(value any) bool {
					return isCompatibleValue(base, value) && !isDefaultValueForType(value)
				})
				if len(defaults) > 0 {
					if len(defaults) == 1 || info.IsSingle() {
						base.Default = defaults[0]
					} else {
						base.Default = defaults
					}
				}
			}

			base.Enum = enumValues
			base.Examples = collect.All(exampleValues, func(value any) bool { return isCompatibleValue(base, value) })
		}
		return item
	})
}

func schemaForProfile(profile config.ProfileInfo) (object *schemaObject) {
	profileType := schemaForPropertySet(profile)
	profileType.Description = "single profile"
	for _, sectionName := range profile.Sections() {
		if info := profile.SectionInfo(sectionName); info != nil {
			profileType.Properties[sectionName] = schemaForPropertySet(info)
		}
	}

	object = newSchemaObject()
	object.Description = "restic profile declarations"
	object.PatternProperties[matchAll] = profileType
	return
}

const (
	version1Pattern = `^(1|)$`
	version2Pattern = `^([2-9]|[1-9][0-9]+)$`
)

func schemaForConfigVersion(version config.Version) SchemaType {
	schema := newSchemaString()
	schema.Description = "configuration format version"
	if version <= config.Version01 {
		schema.Default = "1"
		schema.MaxLength = util.CopyRef(uint64(1))
		schema.Pattern = version1Pattern
	} else {
		schema.Default = "2"
		schema.MinLength = uint64(1)
		schema.Pattern = version2Pattern
	}
	return schema
}

func schemaForGroups(version config.Version) SchemaType {
	info := config.NewGroupInfo()
	object := newSchemaObject()
	object.Description = info.Description()
	var groups SchemaType
	if version <= config.Version01 {
		groups = newSchemaArray(newSchemaString())
		describeAll(groups, "profile-name", "profile names in this group")
	} else {
		groups = schemaForPropertySet(config.NewGroupInfo())
	}
	groups.describe("group", "group declaration")
	object.PatternProperties[matchAll] = groups
	return object
}

func schemaForSchedules() SchemaType {
	info := config.NewScheduleInfo()
	object := newSchemaObject()
	object.Description = info.Description()
	schedules := schemaForPropertySet(info)
	schedules.describe("schedule", "schedule declaration")
	object.PatternProperties[matchAll] = schedules
	return object
}

func schemaForGlobal() SchemaType {
	return schemaForPropertySet(config.NewGlobalInfo())
}

func describeAll(start SchemaType, tile, description string) {
	walkTypes(start, func(item SchemaType) SchemaType {
		if base := item.base(); base != nil && base.Title != "" {
			item.describe(base.Title, description)
		} else {
			item.describe(tile, description)
		}
		return item
	})
}

func schemaForIncludes() SchemaType {
	includesArray := newSchemaArray(newSchemaString())
	includes := newSchemaTypeList(true, newSchemaString(), includesArray) // include or includes-array

	describeAll(includes, "includes", "glob patterns of configuration files to include")
	return includes
}

func schemaForMixins() SchemaType {
	info := config.NewMixinsInfo()
	mixinType := schemaForPropertySet(info)

	object := newSchemaObject()
	object.Description = info.Description()
	object.PatternProperties[matchAll] = mixinType
	return object
}

func schemaForMixinUse() SchemaType {
	info := config.NewMixinUseInfo()
	useType := schemaForPropertySet(info)
	useArray := newSchemaArray(newSchemaTypeList(true, newSchemaString(), useType)) // name or use-object
	useClause := newSchemaTypeList(true, newSchemaString(), useArray)               // name or use-array

	describeAll(useClause, "", info.Description())
	return useClause
}

func applyMixinUseSchema(mixinUseSchema SchemaType, target SchemaType) {
	walkTypes(target, func(item SchemaType) SchemaType {
		if item == mixinUseSchema {
			return nil
		}
		if object, ok := item.(*schemaObject); ok {
			object.Properties[constants.SectionConfigurationMixinUse] = mixinUseSchema
		}
		return item
	})
}

func applyListAppendSchema(target SchemaType) {
	walkTypes(target, func(item SchemaType) SchemaType {
		if object, ok := item.(*schemaObject); ok {
			for name, schemaType := range object.Properties {
				if strings.HasPrefix(name, "...") ||
					strings.HasSuffix(name, "...") ||
					strings.HasSuffix(name, "__APPEND") ||
					strings.HasSuffix(name, "__PREPEND") ||
					name == constants.SectionConfigurationMixinUse {
					continue
				}

				_, isArray := schemaType.(*schemaArray)
				if tl, ok := schemaType.(*schemaTypeList); ok {
					for _, st := range append(tl.OneOf, tl.AnyOf...) {
						if _, isArray = st.(*schemaArray); isArray {
							break
						}
					}
				}

				if isArray {
					object.Properties[name+"__APPEND"] = schemaType
					object.Properties[name+"..."] = schemaType
					object.Properties[name+"__PREPEND"] = schemaType
					object.Properties["..."+name] = schemaType
				}
			}
		}
		return item
	})
}

func schemaForConfigV1(profileInfo config.ProfileInfo) (object *schemaObject) {
	object = schemaForProfile(profileInfo)

	// exclude non-profile properties from profile-schema
	profilesPattern := fmt.Sprintf(`^(?!%s).*$`, strings.Join([]string{
		constants.SectionConfigurationGlobal,
		constants.SectionConfigurationGroups,
		constants.SectionConfigurationIncludes,
		constants.ParameterVersion,
	}, "|"))
	object.PatternProperties[profilesPattern] = object.PatternProperties[matchAll]
	delete(object.PatternProperties, matchAll)

	object.Description = "resticprofile configuration v1"
	object.Properties[constants.SectionConfigurationGlobal] = schemaForGlobal()
	object.Properties[constants.SectionConfigurationGroups] = schemaForGroups(config.Version01)
	object.Properties[constants.SectionConfigurationIncludes] = schemaForIncludes()
	object.Properties[constants.ParameterVersion] = schemaForConfigVersion(config.Version01)
	return
}

func schemaForConfigV2(profileInfo config.ProfileInfo) (object *schemaObject) {
	object = newSchemaObject()
	object.Description = "resticprofile configuration v2"
	object.Properties = map[string]SchemaType{
		constants.SectionConfigurationGlobal:    schemaForGlobal(),
		constants.SectionConfigurationGroups:    schemaForGroups(config.Version02),
		constants.SectionConfigurationIncludes:  schemaForIncludes(),
		constants.SectionConfigurationMixins:    schemaForMixins(),
		constants.SectionConfigurationProfiles:  schemaForProfile(profileInfo),
		constants.SectionConfigurationSchedules: schemaForSchedules(),
		constants.ParameterVersion:              schemaForConfigVersion(config.Version02),
	}
	object.Required = append(object.Required, constants.ParameterVersion)
	{
		mixinUse := schemaForMixinUse()
		for _, section := range []string{
			constants.SectionConfigurationGlobal,
			constants.SectionConfigurationGroups,
			constants.SectionConfigurationProfiles,
		} {
			applyMixinUseSchema(mixinUse, object.Properties[section])
			applyListAppendSchema(object.Properties[section])
		}
	}
	return
}

func generateJsonSchema(version config.Version, resticVersion string) (schema *schemaRoot, err error) {
	var profileInfo config.ProfileInfo
	if len(resticVersion) > 0 {
		profileInfo = config.NewProfileInfoForRestic(resticVersion, true)
	} else {
		profileInfo = config.NewProfileInfo(true)
	}

	var content *schemaObject
	if version <= config.Version01 {
		content = schemaForConfigV1(profileInfo)
	} else {
		content = schemaForConfigV2(profileInfo)
	}

	if schema, err = newSchema(version, resticVersion, content); err == nil {
		const minContentLength = 128 // don't share if content is smaller than potential saving
		schema.createReferences(minContentLength)
	}
	return
}
