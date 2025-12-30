+++
archetype = "chapter"
pre = "<b>3. </b>"
title = "Configuration file"
weight = 3
+++

The configuration file supports various formats:
* [TOML](https://github.com/toml-lang/toml): files with extensions *.toml* and *.conf* (for compatibility with versions before 0.6.0)
* [JSON](https://en.wikipedia.org/wiki/JSON): files with extension *.json*
* [YAML](https://en.wikipedia.org/wiki/YAML): files with extension *.yaml*
* [HCL](https://github.com/hashicorp/hcl): files with extension *.hcl*

* A configuration is a set of [_profiles_]({{% relref "/configuration/profiles" %}}).
* Each profile is in a new section that has the name of the profile.
* Inside each profile, you can specify different flags for each command.
* A command definition is in a subsection of the name of the command.

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
[profile_name]

  [profile_name.backup]

```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
profile_name:

  backup:

```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
profile_name {

    backup = {

    }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "profile_name": {
    "backup": {

    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

All the restic flags can be defined in a section. For most of them you just need to remove the two dashes in front.

To set the flag `--password-file`, the name of the parameter is simply `password-file`.

There's **one exception**: the flag `--repo` is named `repository` in the configuration

## More information

{{% children  %}}
