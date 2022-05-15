+++
chapter = true
pre = "<b>2. </b>"
title = "Configuration"
weight = 10
+++


# Configuration format

* A configuration is a set of _profiles_.
* Each profile is in its own `[section]`.
* Inside each profile, you can specify different flags for each command.
* A command definition is `[section.command]`.

All the restic flags can be defined in a section. For most of them you just need to remove the two dashes in front.

To set the flag `--password-file password.txt` you need to add a line like
```
password-file = "password.txt"
```

There's **one exception**: the flag `--repo` is named `repository` in the configuration

Let's say you normally use this command:

```
restic --repo "local:/backup" --password-file "password.txt" --verbose backup /home
```

For resticprofile to generate this command automatically for you, here's the configuration file:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}

```toml
# indentation is not needed but it makes it easier to read ;)
#
[default]
  repository = "local:/backup"
  password-file = "password.txt"

  [default.backup]
    verbose = true
    source = [ "/home" ]
```

{{% /tab %}}
{{% tab name="yaml" %}}

```yaml
---
default:
  repository: "local:/backup"
  password-file: "password.txt"

  backup:
    verbose: true
    source:
    - "/home"
```

{{% /tab %}}
{{% tab name="hcl" %}}

```hcl

default {
    repository = "local:/backup"
    password-file = "password.txt"

    backup = {
        verbose = true
        source = [ "/home" ]
    }
}
```

{{% /tab %}}
{{% tab name="json" %}}

```json
{
  "default": {
    "repository": "local:/backup",
    "password-file": "password.txt",
    "backup": {
      "verbose": true,
      "source": [
        "/home"
      ]
    }
  }
}
```

{{% /tab %}}
{{% /tabs %}}


You may have noticed the `source` flag is accepting an array of values (inside brackets in TOML, list of values in YAML)

Now, assuming this configuration file is named `profiles.conf` in the current folder, you can simply run

```
resticprofile backup
```

## Path resolution in configuration

All files path in the configuration are resolved from the configuration path. The big **exception** being `source` in `backup` section where it's resolved from the current path where you started resticprofile.
