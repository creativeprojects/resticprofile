---
title: "Upgrade"
weight: 20
---

After installation, upgrade resticprofile to the latest release with this command:

```shell
resticprofile self-update
```

{{% notice style="note" title="Package Managers" %}}
The `self-update` command is generally unavailable when installed through a package manager like Homebrew or Scoop. Use the package manager's upgrade feature instead.
{{% /notice %}}

Resticprofile checks for new versions from GitHub releases and prompts you to update. Use the `-q` or `--quiet` flag to update automatically without prompting.

```shell
resticprofile --quiet self-update
```

or

```shell
resticprofile self-update --quiet
```
