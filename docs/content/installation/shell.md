---
title: "Shell Completion"
weight: 100
---


Shell command line completions are provided for `bash`, `fish`, and `zsh`. 

To load the command completions in shell, use:

```shell
# bash
eval "$(resticprofile generate --bash-completion)"

# fish
resticprofile generate --fish-completion | source

# zsh (add to ~/.zshrc after compinit)
source <(resticprofile generate --zsh-completion)
```

To install them permanently:

```shell
resticprofile generate --bash-completion > /etc/bash_completion.d/resticprofile
chmod +x /etc/bash_completion.d/resticprofile

resticprofile generate --fish-completion > /etc/fish/completions/resticprofile.fish
chmod +x /etc/fish/completions/resticprofile.fish

# zsh: install to a $fpath directory (name must be _resticprofile)
mkdir -p ~/.zsh/completions
resticprofile generate --zsh-completion > ~/.zsh/completions/_resticprofile
# Add to ~/.zshrc before compinit:
# fpath=(~/.zsh/completions $fpath)
# autoload -U compinit && compinit
```
