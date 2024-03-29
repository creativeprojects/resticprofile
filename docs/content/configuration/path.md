---
title: "Path"
date: 2022-04-24T09:44:41+01:00
weight: 10
---

## How paths inside the configuration are resolved

All file paths in the configuration are resolved relative to the configuration path, the path where 
the main configuration file was loaded from. 

The big **exceptions** are `source` in the `backup` section, `status-file`, `prometheus-save-to-file` and the 
restic `repository` (if it is a file). These paths are taken as specified, which means they are resolved 
from the current working directory where you started resticprofile from. 

You can influence this behaviour with profile flag `base-dir`. It allows to set the working directory 
that resticprofile will change into so that profiles do no longer depend on the path where you started 
resticprofile from.

For the following configuration example, when assuming it is stored in `/opt/resticprofile/profiles.*` and 
resticprofile is started from `/home/user/`, the individual paths are resolved to:
* **repository**: `local:/home/user/backup`
* **password-file**: `/opt/resticprofile/password.txt`
* **backup.source**: `/home/user/files`


{{< tabs groupid="config-with-json" >}}
{{% tab title="toml" %}}

```toml
# indentation is not needed but it makes it easier to read ;)
#
version = "1"

[default]
  base-dir = ""
  repository = "local:backup"
  password-file = "password.txt"

  [default.backup]
    source-base = ""
    source = [ "files" ]
```

{{% /tab %}}
{{% tab title="yaml" %}}

```yaml
version: "1"

default:
  base-dir: ""
  repository: "local:backup"
  password-file: "password.txt"

  backup:
    source-base: ""
    source:
    - "files"
```

{{% /tab %}}
{{% tab title="hcl" %}}

```hcl
default {
    base-dir = ""
    repository = "local:backup"
    password-file = "password.txt"

    backup = {
        source-base = ""
        source = [ "files" ]
    }
}
```

{{% /tab %}}
{{% tab title="json" %}}

```json
{
  "version": "1",
  "default": {
    "base-dir": "",
    "repository": "local:backup",
    "password-file": "password.txt",
    "backup": {
      "source-base": "",
      "source": [
        "files"
      ]
    }
  }
}
```

{{% /tab %}}
{{< /tabs >}}

{{% notice hint %}}
Set `base-dir` to an absolute path to resolve `files` and `local:backup` relative to it.
Set `source-base` if you need a separate base path for backup sources.
To set the working directory for the `local:backup` command, use `change-working-dir`. Doing so means that the `source` will not be resolved to an absolute path in this context.
{{% /notice %}}

## How the configuration file is resolved

The default name for the configuration file is `profiles`, without an extension.
You can change the name and its path with the `--config` or `-c` option on the command line.
You can set a specific extension `-c profiles.conf` to load a TOML format file.
If you set a filename with no extension instead, resticprofile will load the first file it finds with any of these extensions:
- .conf (toml format)
- .yaml
- .toml
- .json
- .hcl

### macOS X

resticprofile will search for your configuration file in these folders:
- _current directory_
- ~/Library/Preferences/resticprofile/
- /Library/Preferences/resticprofile/
- /usr/local/etc/
- /usr/local/etc/restic/
- /usr/local/etc/resticprofile/
- /etc/
- /etc/restic/
- /etc/resticprofile/
- /opt/local/etc/
- /opt/local/etc/restic/
- /opt/local/etc/resticprofile/
- ~/ ($HOME directory)

### Other unixes (Linux and BSD)

resticprofile will search for your configuration file in these folders:
- _current directory_
- ~/.config/resticprofile/
- /etc/xdg/resticprofile/
- /usr/local/etc/
- /usr/local/etc/restic/
- /usr/local/etc/resticprofile/
- /etc/
- /etc/restic/
- /etc/resticprofile/
- /opt/local/etc/
- /opt/local/etc/restic/
- /opt/local/etc/resticprofile/
- ~/ ($HOME directory)

### Windows

resticprofile will search for your configuration file in these folders:
- _current directory_
- %USERPROFILE%\AppData\Local\
- c:\ProgramData\
- c:\restic\
- c:\resticprofile\
- %USERPROFILE%\
