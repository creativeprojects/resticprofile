package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// inlineTemplate describes a parsed inline template definition (templates: ...)
type inlineTemplate struct {
	DefaultVariables map[string]interface{} `mapstructure:"default-vars"`
	Source           map[string]interface{} `mapstructure:",remain"`
}

// Resolve applies variables and returns a resolved copy of Source
func (t *inlineTemplate) Resolve(variables map[string]interface{}) map[string]interface{} {
	return t.translate(t.Source, variables)
}

func (t *inlineTemplate) translate(template, variables map[string]interface{}) map[string]interface{} {
	target := make(map[string]interface{})

	for name, rawValue := range template {
		switch value := rawValue.(type) {
		case map[string]interface{}:
			target[name] = t.translate(value, variables)
		case string:
			target[name] = t.expandVariables(value, variables)
		case []interface{}:
			resolved := make([]interface{}, len(value))
			for i := 0; i < len(value); i++ {
				switch item := value[i].(type) {
				case string:
					resolved[i] = t.expandVariables(item, variables)
				case map[string]interface{}:
					resolved[i] = t.translate(item, variables)
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

func (t *inlineTemplate) expandVariables(value string, variables map[string]interface{}) string {
	return os.Expand(value, func(name string) string {
		lookup := strings.ToUpper(name)

		replacement := variables[lookup]
		if replacement == nil {
			replacement = t.DefaultVariables[lookup]
		}

		if replacement != nil {
			if v, err := cast.ToStringE(replacement); err == nil {
				return v
			} else {
				clog.Warningf("unresolved template variable \"%s\": %s", name, err.Error())
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

// parseInlineTemplates returns all valid inline template definitions (templates: ...) from config
func parseInlineTemplates(config *viper.Viper) map[string]*inlineTemplate {
	templates := map[string]*inlineTemplate{}
	definitions := config.GetStringMap(constants.SectionConfigurationTemplates)
	for name, def := range definitions {
		if template, ok := def.(map[string]interface{}); ok {
			it := new(inlineTemplate)
			if err := mapstructure.Decode(template, it); err == nil {
				keysToUpper(it.DefaultVariables)
				templates[name] = it
			} else {
				clog.Warningf("failed parsing \"templates.%s\": %s", name, err.Error())
			}
		} else {
			clog.Warningf("invalid template definition \"templates.%s\" is not an object", name)
		}
	}
	return templates
}

// inlineTemplateCall describes a parsed inline template call (profiles.name.templates: ...)
type inlineTemplateCall struct {
	Name      string                 `mapstructure:"name"`
	Variables map[string]interface{} `mapstructure:"vars"`
}

// parseTemplateCalls parses a template-calls config value (the value of a key with ".templates" suffix)
func parseTemplateCalls(rawValue interface{}) (calls []*inlineTemplateCall, err error) {
	if rawValue != nil {
		switch value := rawValue.(type) {
		case string:
			calls = append(calls, &inlineTemplateCall{Name: value})
		case []interface{}:
			for _, rawItem := range value {
				call := new(inlineTemplateCall)
				calls = append(calls, call)

				switch item := rawItem.(type) {
				case string:
					call.Name = item
				default:
					if err = mapstructure.Decode(item, call); err == nil {
						keysToUpper(call.Variables)
					} else {
						v, _ := stringifyValueOf(item)
						err = fmt.Errorf("cannot parse template call %s: %s", v, err.Error())
						return
					}
				}
			}
		default:
			err = fmt.Errorf("template call must be string or list of strings or list of call objects")
		}
	}

	return
}

// applyInlineTemplates searches for template-calls with suffix ".templates" and applies templates to config
func applyInlineTemplates(config *viper.Viper, templates map[string]*inlineTemplate) (err error) {
	templatesSuffix := fmt.Sprintf(".%s", constants.SectionConfigurationTemplates)

	for _, key := range config.AllKeys() {
		if !strings.HasSuffix(key, templatesSuffix) {
			continue
		}

		var calls []*inlineTemplateCall
		if calls, err = parseTemplateCalls(config.Get(key)); err == nil {
			config.Set(key, nil)
			configKey := strings.TrimSuffix(key, templatesSuffix)

			if targetConfig := config.Sub(configKey); targetConfig != nil {
				for _, call := range calls {
					if template, found := templates[call.Name]; found {
						templateMap := template.Resolve(call.Variables)
						revolveAppendToListKeys(targetConfig, templateMap)
						err = targetConfig.MergeConfigMap(templateMap)
					} else {
						err = fmt.Errorf("undefined template \"%s\"", call.Name)
					}

					if err != nil {
						break
					}
				}
			}
		}

		if err != nil {
			err = fmt.Errorf("failed applying templates %s: %s", key, err.Error())
			return
		}
	}
	return
}

const (
	KEY_SUFFIX_APPEND_FIRST = "+0"
	KEY_SUFFIX_APPEND_LAST  = "++"
)

// revolveAppendToListKeys resolves "key++" and "key+0" in template using config as source
func revolveAppendToListKeys(config *viper.Viper, template map[string]interface{}) {
	for name, value := range template {
		first := strings.HasSuffix(name, KEY_SUFFIX_APPEND_FIRST)
		last := strings.HasSuffix(name, KEY_SUFFIX_APPEND_LAST)

		if !first && !last {
			if child, ok := value.(map[string]interface{}); ok {
				var cc *viper.Viper
				if config != nil {
					cc = config.Sub(name)
				}
				revolveAppendToListKeys(cc, child)
			}
			continue
		} else {
			delete(template, name)
		}

		targetName := name
		targetName = strings.TrimSuffix(targetName, KEY_SUFFIX_APPEND_FIRST)
		targetName = strings.TrimSuffix(targetName, KEY_SUFFIX_APPEND_LAST)

		var source, appendable []interface{}
		if config != nil {
			sourceValue := config.Get(targetName)
			if source = cast.ToSlice(sourceValue); source == nil && sourceValue != nil {
				source = []interface{}{sourceValue}
			}
		}

		if appendable = cast.ToSlice(value); appendable == nil && value != nil {
			appendable = []interface{}{value}
		}

		if first {
			template[targetName] = append(appendable, source...)
		} else if last {
			template[targetName] = append(source, appendable...)
		}
	}
}
