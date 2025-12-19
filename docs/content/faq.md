---
title: FAQ
pre: "<b>4. </b>"
weight: 4
---

{{< toc >}}


## Installation
### Shell Completion

To generate the shell completion script, use:

```shell
# bash
eval "$(resticprofile generate --bash-completion)"

# zsh
eval "$(resticprofile generate --zsh-completion)"
```

### My homebrew install failed!

Homebrew appears to need access to a compiler. If the installation fails, you may need to install gcc.
