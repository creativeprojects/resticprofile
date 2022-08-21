+++
chapter = true
pre = "<b>2. </b>"
title = "Configuration"
weight = 10
+++


# Configuration file

* A configuration is a set of _profiles_.
* Each profile is in its own `[section]` (in TOML).
* Inside each profile, you can specify different flags for each command.
* A command definition is `[section.command]` (in TOML).

All the restic flags can be defined in a section. For most of them you just need to remove the two dashes in front.

To set the flag `--password-file`, the name of the parameter is simply `password-file`.

There's **one exception**: the flag `--repo` is named `repository` in the configuration

## Example 

So let's say you normally use this simple command:

```shell
$ restic --repo "local:/backup" --password-file "password.txt" --verbose backup /home
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

Now, assuming this configuration file is named `profiles.conf` in the current folder (it's the default config file name), you can simply run

```shell
$ resticprofile backup
```

and resticprofile will do its magic and generate the command line for you.

If you have any doubt on what it's running, you can try a `--dry-run`:

```shell
$ resticprofile --dry-run backup
2022/05/18 17:14:07 profile 'default': starting 'backup'
2022/05/18 17:14:07 dry-run: /usr/bin/restic backup --password-file key --repo local:/backup --verbose /home
2022/05/18 17:14:07 profile 'default': finished 'backup'
```

## Path resolution in configuration

All files path in the configuration are resolved from the configuration path. The big **exception** being `source` in `backup` section where it's resolved from the current path where you started resticprofile.

Using the basic configuration from earlier, and taking into account that the configuration file is saved in the directory `/opt/resticprofile`, the password key file `password.txt` is expected to be found at `/opt/resticprofile/password.txt` no matter your current directory.

## More information

{{% children  %}}
