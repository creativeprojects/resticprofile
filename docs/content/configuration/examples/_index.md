---
title: "Examples"
weight: 40
alwaysopen: false
---

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
{{< /tabs >}}


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

## More Examples

{{% children sort="weight" depth="2" %}}
