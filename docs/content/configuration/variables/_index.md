---
title: "Variables"
weight: 31
alwaysopen: false
---

{{< toc >}}

## Variable expansion in configuration file

You might want to reuse the same configuration (or bits of it) on different environments. One way of doing it is to create a generic configuration where specific bits can be replaced by a variable.

There are two kinds of variables:
- **template variables**: These variables are fixed once the full configuration file is loaded: [includes]({{% relref "/configuration/profiles/include" %}}) are loaded, and [inheritance]({{% relref "/configuration/profiles/inheritance" %}}) is resolved. These variables are replaced by their value **before** the configuration is parsed.
- **runtime variables**: These variables are replaced by their value **after** the configuration is parsed. In other words: these variables are replaced by their value just before the command is executed.

## Template variables

The syntax for using a pre-defined variable is:

```
{{ .VariableName }}
```

The list of pre-defined variables and environment variables can be found in the [reference]({{% relref "/configuration/variables/env_vars" %}})

Environment variables are accessible using `.Env.` followed by the (upper case) name of the environment variable.

Default and fallback values for an empty or unset variable can be declared with `{{ ... | or ... }}`.
For example `{{ .Env.HOME | or .Env.USERPROFILE | or "/fallback-homedir" }}` will try to resolve `$HOME`, if empty try to resolve `$USERPROFILE` 
and finally default to `/fallback-homedir` if none of the env variables are defined.

The variables `.OS` and `.Arch` are filled with the target platform that `resticprofile` was compiled for (see 
[releases](https://github.com/creativeprojects/resticprofile/releases) for more information on existing precompiled platform binaries). 

For variables that are objects, you can call all public fields or methods on it.
For example, for the variable `.Now` ([time.Time](https://golang.org/pkg/time/)) you can use:

- `(.Now.AddDate years months days)`
- `.Now.Day`
- `.Now.Format layout`
- `.Now.Hour`
- `.Now.Minute`
- `.Now.Month`
- `.Now.Second`
- `.Now.UTC`
- `.Now.Unix`
- `.Now.Weekday`
- `.Now.Year`
- `.Now.YearDay`

Time can be formatted with `.Now.Format layout`, for example `{{ .Now.Format "2006-01-02T15:04:05Z07:00" }}` formats the current time as RFC3339 timestamp. 
Check [time.Time#constants](https://pkg.go.dev/time#pkg-constants) for more layout examples.

The variable `.Now` also allows to derive a relative `Time`. For example `{{ (.Now.AddDate 0 -6 -14).Format "2006-01-02" }}` formats a date that 
is 6 months and 14 days before now.

## Hand-made variables

You can also define variables yourself. Hand-made variables starts with a `$` ([PHP](https://en.wikipedia.org/wiki/PHP) anyone?) and get declared and assigned with the `:=` operator ([Pascal](https://en.wikipedia.org/wiki/Pascal_(programming_language)) anyone?).

{{% notice style="info" %}}
You can only use double quotes `"` to declare the string, single quotes `'` are not allowed. You can also use backticks to declare the string.
{{% /notice %}}

Here's an example:

```yaml
# declare and assign a value to the variable
{{ $name := "something" }}

profile:
  # put the content of the variable here
  tag: "{{ $name }}"
```
{{% notice style="note" %}}
Variables are only valid in the file they are declared in. They cannot be shared in files loaded via `include`.
{{% /notice %}}

Variables can be redefined using the `=` operator. The new value will be used from the point of redefinition to the end of the file.

```yaml
# declare and assign a value to the variable
{{ $name := "something" }}

# reassign a new value to the variable
{{ $name = "something else" }}

```

### Windows path inside a variable

Windows path are using backslashes `\` and are interpreted as escape characters in the configuration file. To use a Windows path inside a variable, you have a few options:
- you can escape the backslashes with another backslash.
- you can use forward slashes `/` instead of backslashes. Windows is able to use forward slashes in paths.
- you can use the backtick to declare the string instead of a double quote.

For example:
```yaml
# double backslash
{{ $path := "C:\\Users\\CP\\Documents" }}
# forward slash
{{ $path := "C:/Users/CP/Documents" }}
# backticks
{{ $path := `C:\Users\CP\Documents` }}
```

## Runtime variable expansion

Variable expansion as described in the previous section using the `{{ .Var }}` syntax refers to [template variables]({{% relref "/configuration/variables/templates" %}}) that are expanded prior to parsing the configuration file. 
This means they must be used carefully to create correct config markup, but they are also very flexible.

There is also unix style variable expansion using the `${variable}` or `$variable` syntax on configuration **values** that expand after the config file was parsed. Values that take a file path or path expression and a few others support this expansion. 

If not specified differently, these variables resolve to the corresponding environment variable or to an empty value if no such environment variable exists. Exceptions are [mixins]({{% relref "/configuration/profiles/inheritance#mixins" %}}) where `$variable` style is used for parametrisation and the profile [config flag]({{% relref "reference/profile" %}}) `prometheus-push-job`.

{{% notice style="tip" %}}
Use `$$` to escape a single `$` in configuration values that support variable expansion. E.g. on Windows you might want to exclude `$RECYCLE.BIN`. Specify it as: `exclude = ["$$RECYCLE.BIN"]`.
{{% /notice %}}
