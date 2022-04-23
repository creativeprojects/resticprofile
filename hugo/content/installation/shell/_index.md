---
title: "Shell Completion"
date: 2022-04-23T23:55:10+01:00
weight: 100
---


Shell command line completions are provided for `bash` and `zsh`. 

To load the command completions in shell, use:

```shell
# bash
eval "$(resticprofile completion-script --bash)"

# zsh
eval "$(resticprofile completion-script --zsh)"
```

To install them permanently:

```
$ resticprofile completion-script --bash > /etc/bash_completion.d/resticprofile
$ chmod +x /etc/bash_completion.d/resticprofile
```
