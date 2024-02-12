package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/maybe"
)

// resticDescriptionReplacements removes or replaces misleading documentation fragments from restic man pages
var resticDescriptionReplacements = []struct {
	from *regexp.Regexp
	to   string
}{
	{
		from: regexp.MustCompile(`\(specify multiple times or a level using -+verbose=n, max level/times (is \d+)\)`),
		to:   `(true for level 1 or a number for increased verbosity, max level $1)`,
	},
	{
		from: regexp.MustCompile(`\s*\(can be (specified|given) multiple times\)\s*(\.|$)`),
		to:   `$2`,
	},
	{
		from: regexp.MustCompile(`[;,]\s+can be (specified|given) multiple times\s*\)\s*(\.|$)`),
		to:   `)$2`,
	},
}

func resticDescriptionFilter(description string) string {
	for _, replacement := range resticDescriptionReplacements {
		if replacement.from.MatchString(description) {
			description = replacement.from.ReplaceAllString(description, replacement.to)
		}
	}
	return description
}

func init() {
	// Restic Descriptions: Remove or replace misleading documentation fragments
	registerPropertyInfoCustomizer(func(sectionName, propertyName string, property accessibleProperty) {
		if property.option() != nil { // only restic options
			property.basic().addDescriptionFilter(resticDescriptionFilter)
		}
	})

	// Global: "default-command" add all restic commands as example values
	registerPropertyInfoCustomizer(func(sectionName, propertyName string, property accessibleProperty) {
		if sectionName == constants.SectionConfigurationGlobal && propertyName == "default-command" {
			property.basic().examples = restic.CommandNamesForVersion(restic.AnyVersion)
		}
	})

	// Profile: special handling for verbose (can be bool or int, detection catches bool only)
	registerPropertyInfoCustomizer(func(sectionName, propertyName string, property accessibleProperty) {
		if propertyName == constants.ParameterVerbose && property.option() != nil {
			info := property.basic()
			info.mayBool = true
			info.mayNumber = true
			info.mustInt = true
		}
	})

	// Profile: special handling for host, path and tag
	registerPropertyInfoCustomizer(func(sectionName, propertyName string, property accessibleProperty) {
		if propertyName == constants.ParameterHost || propertyName == constants.ParameterPath || propertyName == constants.ParameterTag {
			note := ""
			info := property.basic()
			info.mayBool = true
			info.examples = []string{"true", "false", fmt.Sprintf(`"%s"`, propertyName)}

			suffixDefaultTrueV2 := fmt.Sprintf(` Defaults to true for config version 2 in "%s".`, sectionName)

			if propertyName == constants.ParameterHost {
				info.format = "hostname"
				note = `Boolean true is replaced with the hostname of the system.`
			} else {
				note = fmt.Sprintf(`Boolean true is replaced with the %ss from section "backup".`, propertyName)
			}

			if sectionName == constants.CommandBackup {
				if propertyName != constants.ParameterHost {
					info.examples = info.examples[1:] // remove "true" from examples of backup section
					note = fmt.Sprintf(`Boolean true is unsupported in section "backup".`)
				} else {
					note += suffixDefaultTrueV2
				}
			} else if sectionName == constants.SectionConfigurationRetention {
				if propertyName == constants.ParameterHost {
					note = `Boolean true is replaced with the hostname that applies in section "backup".`
				}
				if propertyName == constants.ParameterPath {
					note += ` Defaults to true in "retention".`
				} else {
					note += suffixDefaultTrueV2
				}
			}

			if note != "" {
				info.addDescriptionFilter(func(desc string) string {
					return fmt.Sprintf("%s.\n%s", strings.TrimSuffix(strings.TrimSpace(desc), "."), note)
				})
			}
			return
		}
	})

	// Profile: special handling for ConfidentialValue
	confidentialType := reflect.TypeOf(ConfidentialValue{})
	registerPropertyInfoCustomizer(func(sectionName, propertyName string, property accessibleProperty) {
		if field := property.field(); field != nil {
			if util.ElementType(field.Type).AssignableTo(confidentialType) {
				basic := property.basic().resetTypeInfo()
				if field.Type.Kind() == reflect.Map {
					basic.nested = &namedPropertySet{name: propertyName, propertySet: propertySet{openSet: true}}
				} else {
					basic.mayString = true
				}
			}
		}
	})

	// Profile: special handling for maybe.Bool
	maybeBoolType := reflect.TypeOf(maybe.Bool{})
	registerPropertyInfoCustomizer(func(sectionName, propertyName string, property accessibleProperty) {
		if field := property.field(); field != nil {
			if util.ElementType(field.Type).AssignableTo(maybeBoolType) {
				basic := property.basic().resetTypeInfo()
				basic.mayBool = true
			}
		}
	})

	// Profile: deprecated sections (squash with deprecated, e.g. schedule in retention)
	registerPropertyInfoCustomizer(func(sectionName, propertyName string, property accessibleProperty) {
		if field := property.sectionField(nil); field != nil {
			if _, deprecated := field.Tag.Lookup("deprecated"); deprecated {
				property.basic().deprecated = true
			}
		}
	})

	// Profile: exclude "help" (--help flag doesn't need to be in the reference)
	ExcludeProfileProperty("*", "help")
}
