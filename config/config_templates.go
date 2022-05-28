package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type inlineTemplate struct {
	source map[string]interface{}
}

func (t *inlineTemplate) toMap(variables map[string]string) map[string]interface{} {
	return t.translate(t.source, variables)
}

func (t *inlineTemplate) expandVariables(value string, variables map[string]string) string {
	return os.Expand(value, func(name string) string {
		if replacement, ok := variables[name]; ok {
			return replacement
		} else {
			// retain unknown variables (could be env vars to be resolved later)
			return fmt.Sprintf("${%s}", name)
		}
	})
}

func (t *inlineTemplate) translate(template map[string]interface{}, params map[string]string) map[string]interface{} {
	target := map[string]interface{}{}

	for name, rawValue := range template {
		switch value := rawValue.(type) {
		case map[string]interface{}:
			target[name] = t.translate(value, params)
		case string:
			target[name] = t.expandVariables(value, params)
		case []interface{}:
			resolved := make([]interface{}, len(value))
			for i := 0; i < len(value); i++ {
				switch value[i].(type) {
				case string:
					resolved[i] = t.expandVariables(value[i].(string), params)
				default:
					resolved[i] = value[i]
				}
			}
			target[name] = resolved
		default:
			target[name] = value
		}
	}

	return target
}

// parseInlineTemplates returns any defined inline template definitions (templates: ...)
func parseInlineTemplates(config *viper.Viper) map[string]*inlineTemplate {
	templates := map[string]*inlineTemplate{}
	definitions := config.GetStringMap(constants.SectionConfigurationTemplates)
	for name, def := range definitions {
		if template, ok := def.(map[string]interface{}); ok {
			templates[name] = &inlineTemplate{source: template}
		} else {
			clog.Warningf("invalid template definition \"templates.%s\" is not an object", name)
		}
	}
	return templates
}

type inlineTemplateCall struct {
	Name      string            `mapstructure:"name"`
	Variables map[string]string `mapstructure:"vars"`
}

// parseTemplateCalls finds all template calls in the value of a config key with ".templates" suffix
func parseTemplateCalls(rawValue interface{}) (calls []*inlineTemplateCall, err error) {
	if !reflect.ValueOf(rawValue).IsNil() {
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
					if err = mapstructure.Decode(item, call); err != nil {
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

// applyInlineTemplates looks for keys suffixed with ".templates" and applies matching templates
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
						err = targetConfig.MergeConfigMap(template.toMap(call.Variables))
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
