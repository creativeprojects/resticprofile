---
title: "Shell Completion"
weight: 100
---


Shell command line completions are provided for `bash`, `fish`, and `zsh`.

The completion scripts are built into resticprofile and printed by the `generate` command, so they always match the version you have installed.

## Bash

To load the completions in the current shell:

```shell
eval "$(resticprofile generate --bash-completion)"
```

To install them permanently:

```shell
resticprofile generate --bash-completion > /etc/bash_completion.d/resticprofile
chmod +x /etc/bash_completion.d/resticprofile
```

## Fish

To load the completions in the current shell:

```shell
resticprofile generate --fish-completion | source
```

To install them permanently:

```shell
resticprofile generate --fish-completion > /etc/fish/completions/resticprofile.fish
chmod +x /etc/fish/completions/resticprofile.fish
```

## Zsh

The zsh completion proposes resticprofile commands and flags, and the profile and group names found in your configuration file. A description is shown next to each match:

```
$ resticprofile default.<TAB>
default.profiles    -- display profile names from the configuration file
default.schedule    -- schedule jobs from a profile or group (or of all profiles and groups)
default.show        -- show all the details of the current profile or group
default.status      -- display the status of scheduled jobs of a profile or group (or of all profiles and groups)
default.unschedule  -- remove scheduled jobs of a profile or group (or of all profiles and groups)
```

It also completes restic commands and flags when the [restic completion](#completing-restic-commands) is installed.

### Requirements

The zsh completion system (`compinit`) must be enabled. Frameworks like Oh My Zsh enable it for you. If you don't use one, make sure your `~/.zshrc` contains:

```zsh
autoload -Uz compinit && compinit
```

### Try it in the current shell

```zsh
source <(resticprofile generate --zsh-completion)
```

The completions are available until you close the shell. To keep this method, add the line to your `~/.zshrc`, **after** `compinit`.

### Install permanently

zsh loads its completions from the directories listed in `$fpath`. Save the script in one of these directories, in a file named `_resticprofile`. The leading underscore is required: zsh ignores a file named `resticprofile`. The file doesn't need to be executable.

{{< tabs groupid="zsh-completion-install" >}}
{{% tab title="Homebrew" %}}

Homebrew installs the bash, fish and zsh completions along with resticprofile. There's nothing else to do, as long as [Homebrew completions](https://docs.brew.sh/Shell-Completion) are enabled in your shell.

{{% /tab %}}
{{% tab title="Oh My Zsh" %}}

Oh My Zsh adds the `completions` directory of its custom folder to `$fpath`, and runs `compinit` itself:

```zsh
mkdir -p "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/completions"
resticprofile generate --zsh-completion > "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/completions/_resticprofile"
```

There's nothing to add to your `~/.zshrc`.

{{% /tab %}}
{{% tab title="Your own directory" %}}

```zsh
mkdir -p ~/.zsh/completions
resticprofile generate --zsh-completion > ~/.zsh/completions/_resticprofile
```

Then add the directory to `$fpath` in your `~/.zshrc`, **before** `compinit`:

```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

{{% /tab %}}
{{% tab title="All users" %}}

`/usr/local/share/zsh/site-functions` is in the default `$fpath` on most systems (check with `print -l $fpath`):

```zsh
sudo mkdir -p /usr/local/share/zsh/site-functions
resticprofile generate --zsh-completion | sudo tee /usr/local/share/zsh/site-functions/_resticprofile > /dev/null
```

There's nothing to add to your `~/.zshrc`.

{{% /tab %}}
{{< /tabs >}}

Start a new shell (or run `exec zsh`) to load the completions.

### Completing restic commands

resticprofile hands the completion of restic commands and flags (like `backup` or `--host`) over to the zsh completion provided by restic. Without it, only the resticprofile commands, flags and profile names are completed.

Packages of restic usually install it. If it's missing, save it next to the resticprofile one, in a file named `_restic`:

```zsh
restic generate --zsh-completion ~/.zsh/completions/_restic
```

### Troubleshooting

If pressing `TAB` proposes file names instead of resticprofile commands, check that zsh knows about the completion:

```zsh
print $_comps[resticprofile]
```

This should print `_resticprofile`. If it prints nothing:

- the file must be named `_resticprofile`, with the leading underscore
- its directory must be listed in `$fpath` (check with `print -l $fpath`), and added **before** `compinit` runs
- when using `source` instead, the line must come **after** `compinit`: zsh reports `command not found: compdef` otherwise
- zsh caches the list of completions: delete the cache with `rm -f ~/.zcompdump*` and start a new shell
