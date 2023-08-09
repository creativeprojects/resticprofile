+++
archetype = "chapter"
pre = "<b>2. </b>"
title = "Configuration file"
weight = 2
+++


* A configuration is a set of _profiles_.
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
{{% /tabs %}}

All the restic flags can be defined in a section. For most of them you just need to remove the two dashes in front.

To set the flag `--password-file`, the name of the parameter is simply `password-file`.

There's **one exception**: the flag `--repo` is named `repository` in the configuration

## Example 

So let's say you normally use this simple command:

```shell
restic --repo "local:/backup" --password-file "password.txt" --verbose backup /home
```

For resticprofile to generate this command automatically for you, here's the configuration file:

{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
# indentation is not needed but it makes it easier to read ;)
#
version = "1"

[default]
  repository = "local:/backup"
  password-file = "password.txt"

  [default.backup]
    verbose = true
    source = [ "/home" ]
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

default:
  repository: "local:/backup"
  password-file: "password.txt"

  backup:
    verbose: true
    source:
    - "/home"
```

{{% /tab %}}
{{% tab title="hcl" %}}

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
{{% tab title="json" %}}

```json
{
  "version": "1",
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
resticprofile backup
```

and resticprofile will do its magic and generate the command line for you.

If you have any doubt on what it's running, you can try a `--dry-run`:

```shell
resticprofile --dry-run backup
2022/05/18 17:14:07 profile 'default': starting 'backup'
2022/05/18 17:14:07 dry-run: /usr/bin/restic backup --password-file password.txt --repo local:/backup --verbose /home
2022/05/18 17:14:07 profile 'default': finished 'backup'
```

## More information

{{% children  %}}
