---
title: "Linux and MacOS"
weight: 10
---

{{< toc >}}

## Installation via a script
Works for: MacOS, Linux


Here's a simple script to download the binary in a `bin` directory under your current directory.

```shell
curl -sfL https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh | sh
```

If you need more control, you can save the shell script and run it manually:

```shell
curl -LO https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
chmod +x install.sh
sudo ./install.sh -b /usr/local/bin
```

It will install resticprofile in `/usr/local/bin/`


## Installation With Homebrew
Works for: MacOS, Linux

There's a [homebrew](https://brew.sh/) tap for resticprofile:

```shell
brew tap creativeprojects/tap
brew install resticprofile
# You can also install `restic` at the same time with `--with-restic` flag:
brew install resticprofile --with-restic
```

{{% notice style="note" %}}
The resticprofile command `self-update` is not available when installed via homebrew.
{{% /notice %}}

{{% notice style="tip" title="Debian package" %}}
WakeMeOps (a third party) publishes packages for [restic](https://docs.wakemeops.com/packages/restic/) and [resticprofile](https://docs.wakemeops.com/packages/resticprofile/).
{{% /notice %}}
