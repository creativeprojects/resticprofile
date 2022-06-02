package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// mixin describes a parsed mixin definition (mixins: ...)
type mixin struct {
	DefaultVariables map[string]interface{} `mapstructure:"default-vars"`
	Source           map[string]interface{} `mapstructure:",remain"`
}

// Resolve applies variables and returns a resolved copy of Source
func (m *mixin) Resolve(variables map[string]interface{}) map[string]interface{} {
	return m.translate(m.Source, variables)
}

func (m *mixin) translate(source, variables map[string]interface{}) map[string]interface{} {
	target := make(map[string]interface{})

	for name, rawValue := range source {
		switch value := rawValue.(type) {
		case map[string]interface{}:
			target[name] = m.translate(value, variables)
		case string:
			target[name] = m.expandVariables(value, variables)
		case []interface{}:
			resolved := make([]interface{}, len(value))
			for i := 0; i < len(value); i++ {
				switch item := value[i].(type) {
				case string:
					resolved[i] = m.expandVariables(item, variables)
				case map[string]interface{}:
					resolved[i] = m.translate(item, variables)
				default:
					resolved[i] = item
				}
			}
			target[name] = resolved
		default:
			target[name] = value
		}
	}

	return target
}

func (m *mixin) expandVariables(value string, variables map[string]interface{}) string {
	return os.Expand(value, func(name string) string {
		lookup := strings.ToUpper(name)

		replacement := variables[lookup]
		if replacement == nil {
			replacement = m.DefaultVariables[lookup]
		}

		if replacement != nil {
			if v, err := cast.ToStringE(replacement); err == nil {
				return v
			} else {
				clog.Warningf("unresolved mixin variable \"%s\": %s", name, err.Error())
			}
		}

		// retain unknown variables (could be env vars to be resolved later)
		return fmt.Sprintf("${%s}", name)
	})
}

func keysToUpper(items map[string]interface{}) map[string]interface{} {
	for key, value := range items {
		if lk := strings.ToUpper(key); lk != key {
			delete(items, key)
			items[lk] = value
		}
	}
	return items
}

// parseMixins returns all valid mixin definitions (mixins: ...) from config
func parseMixins(config *viper.Viper) map[string]*mixin {
	mixins := map[string]*mixin{}
	definitions := config.GetStringMap(constants.SectionConfigurationMixins)
	for name, def := range definitions {
		if definition, ok := def.(map[string]interface{}); ok {
			{
				buf := &strings.Builder{}
				yaml.NewEncoder(buf).Encode(definition)
				clog.Tracef("mixin declaration \"%s\": \n%s", name, buf.String())
			}
			mi := new(mixin)
			if err := mapstructure.Decode(definition, mi); err == nil {
				keysToUpper(mi.DefaultVariables)
				mixins[name] = mi
			} else {
				clog.Warningf("failed parsing \"mixins.%s\": %s", name, err.Error())
			}
		} else {
			clog.Warningf("invalid mixin definition \"mixins.%s\" is not an object", name)
		}
	}
	return mixins
}

// mixinUse the use of a mixin within the configuration (profiles.name.use: ...)
type mixinUse struct {
	Name              string                 `mapstructure:"name"`
	Variables         map[string]interface{} `mapstructure:"vars"`
	ImplicitVariables map[string]interface{} `mapstructure:",remain"`
}

func (u *mixinUse) normalizeVariables() {
	keysToUpper(u.Variables)

	if len(u.ImplicitVariables) > 0 {
		keysToUpper(u.ImplicitVariables)
		if u.Variables == nil {
			u.Variables = u.ImplicitVariables
		} else {
			for k, v := range u.ImplicitVariables {
				if _, exists := u.Variables[k]; !exists {
					u.Variables[k] = v
				}
			}
		}
		u.ImplicitVariables = nil
	}
}

// parseMixinUses parses a mixin use config value (the value of a key with ".use" suffix)
func parseMixinUses(rawValue interface{}) (uses []*mixinUse, err error) {
	if rawValue != nil {
		switch value := rawValue.(type) {
		case string:
			uses = append(uses, &mixinUse{Name: value})
		case []interface{}:
			for _, rawItem := range value {
				use := new(mixinUse)
				uses = append(uses, use)

				switch item := rawItem.(type) {
				case string:
					use.Name = item
				default:
					if err = mapstructure.Decode(item, use); err == nil {
						use.normalizeVariables()
					} else {
						v, _ := stringifyValueOf(item)
						err = fmt.Errorf("cannot parse mixin use %s: %s", v, err.Error())
						return
					}
				}
			}
		default:
			err = fmt.Errorf("mixin use must be string or list of strings or list of use objects")
		}
	}

	return
}

func mergeConfigMap(config *viper.Viper, configKey, keyDelimiter string, content map[string]interface{}) error {
	path := strings.Split(configKey, keyDelimiter)
	for i := len(path) - 1; i >= 0; i-- {
		container := make(map[string]interface{})
		container[path[i]], content = content, container
	}
	return config.MergeConfigMap(content)
}

// applyMixins applies mixins to config where they are referenced with "use" keys
func applyMixins(config *viper.Viper, keyDelimiter string, mixins map[string]*mixin) (err error) {
	useSuffix := keyDelimiter + constants.SectionConfigurationMixinUse

	for _, key := range config.AllKeys() {
		if !strings.HasSuffix(key, useSuffix) {
			continue
		}

		var uses []*mixinUse
		if uses, err = parseMixinUses(config.Get(key)); err == nil {
			configKey := strings.TrimSuffix(key, useSuffix)

			for _, use := range uses {
				if mi, found := mixins[use.Name]; found {
					content := mi.Resolve(use.Variables)
					revolveAppendToListKeys(config.Sub(configKey), content)
					err = mergeConfigMap(config, configKey, keyDelimiter, content)
				} else {
					err = fmt.Errorf("undefined mixin \"%s\"", use.Name)
				}

				if err != nil {
					break
				}
			}
		}

		if err != nil {
			err = fmt.Errorf("failed applying %s: %s", strings.ReplaceAll(key, keyDelimiter, "."), err.Error())
			return
		}
	}
	return
}

type mixinAppendToKeyType int

const (
	mixinNoAppend mixinAppendToKeyType = iota
	mixinAppend
	mixinPrepend
)

var (
	mixinPrependKeyRegex = regexp.MustCompile(`(?i)^(.+)__PREPEND$|^\.\.\.(.+)$`)
	mixinAppendKeyRegex  = regexp.MustCompile(`(?i)^(.+)__APPEND$|^(.+)\.\.\.$`)
)

func parseAppendToListKey(key string) (targetKey string, operation mixinAppendToKeyType) {
	var match []string
	if match = mixinAppendKeyRegex.FindStringSubmatch(key); len(match) > 1 {
		operation = mixinAppend
	} else if match = mixinPrependKeyRegex.FindStringSubmatch(key); len(match) > 1 {
		operation = mixinPrepend
	}

	if len(match) > 1 {
		for _, targetKey = range match[1:] {
			if len(targetKey) > 0 {
				break
			}
		}
	}

	if len(targetKey) == 0 {
		operation = mixinNoAppend
	}

	return
}

// revolveAppendToListKeys resolves "key__APPEND" and "key__PREPEND" in content using config as base
func revolveAppendToListKeys(config *viper.Viper, content map[string]interface{}) {
	for name, value := range content {
		targetName, operation := parseAppendToListKey(name)

		if operation == mixinNoAppend {
			if child, ok := value.(map[string]interface{}); ok {
				var cc *viper.Viper
				if config != nil {
					cc = config.Sub(name)
				}
				revolveAppendToListKeys(cc, child)
			}
			continue
		} else {
			delete(content, name)
		}

		sourceValue := content[targetName]
		if sourceValue == nil && config != nil {
			sourceValue = config.Get(targetName)
		}

		var source, appendable []interface{}
		if source = cast.ToSlice(sourceValue); source == nil && sourceValue != nil {
			source = []interface{}{sourceValue}
		}

		if appendable = cast.ToSlice(value); appendable == nil && value != nil {
			appendable = []interface{}{value}
		}

		switch operation {
		case mixinAppend:
			content[targetName] = append(source, appendable...)
		case mixinPrepend:
			content[targetName] = append(appendable, source...)
		}
	}
}
