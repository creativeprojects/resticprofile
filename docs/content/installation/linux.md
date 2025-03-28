---
title: "Linux"
weight: 10
---


## Installation via a script

Here's a simple script to download the binary automatically. It works on macOS, FreeBSD and Linux:

```shell
curl -sfL https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh | sh
```

It should copy resticprofile in a `bin` directory under your current directory.

If you need more control, you can save the shell script and run it manually:

```shell
curl -LO https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
chmod +x install.sh
sudo ./install.sh -b /usr/local/bin
```

It will install resticprofile in `/usr/local/bin/`

## Debian package

WakeMeOps publishes packages for [restic](https://docs.wakemeops.com/packages/restic/) and [resticprofile](https://docs.wakemeops.com/packages/resticprofile/).

Learn more: https://docs.wakemeops.com/packages/resticprofile/

These packages are managed by a third party.

## Installation with homebrew for Linux

There's a Linux [homebrew](https://brew.sh/) tap for resticprofile:

```shell
brew tap creativeprojects/tap
brew install resticprofile
```

You can also install `restic` at the same time with `--with-restic` flag:

```shell
brew install resticprofile --with-restic
```

You can test that resticprofile is properly installed (make sure you have restic installed or the test will fail):

```shell
brew test resticprofile
```

Upgrading resticprofile installed via homebrew is very easy:

```shell
brew update
brew upgrade resticprofile
```

{{% notice style="note" %}}
The resticprofile command `self-update` is not available when installed via homebrew.
{{% /notice %}}

### Note on installing on Linux via Homebrew

When testing homebrew after spinning new Linux virtual machines, I noticed resticprofile wouldn't install without a compiler installed on the machine.
Even though resticprofile is distributed as a **binary**, it looks like homebrew needs access to a compiler.

Depending on your distribution you will need to install gcc:
* `sudo yum install gcc`
* `sudo apt install gcc`
