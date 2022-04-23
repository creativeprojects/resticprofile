---
title: "mac OS X"
date: 2022-04-23T23:22:41+01:00
weight: 11
---

## Installation with Homebrew

There's a [homebrew](https://brew.sh/) tap for resticprofile:

```
$ brew tap creativeprojects/tap
$ brew install resticprofile
```

You can also install `restic` at the same time with `--with-restic` flag:

```
$ brew install resticprofile --with-restic
```

You can test that resticprofile is properly installed (make sure you have restic installed or the test will fail):

```
$ brew test resticprofile
```

Upgrading resticprofile installed via homebrew is very easy:

```
$ brew update
$ brew upgrade resticprofile
```

## Installation via a script

Here's a simple script to download the binary automatically. It works on mac OS X, FreeBSD and Linux:

```
$ curl -sfL https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh | sh
```

It should copy resticprofile in a `bin` directory under your current directory.

If you need more control, you can save the shell script and run it manually:

```
$ curl -LO https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
$ chmod +x install.sh
$ sudo ./install.sh -b /usr/local/bin
```

It will install resticprofile in `/usr/local/bin/`
