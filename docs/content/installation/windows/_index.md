---
title: "Windows"
date: 2022-04-23T23:24:38+01:00
weight: 12
---



## Installation using bash

You can use a script if you're using bash in Windows (via WSL, git bash, etc.)

```shell
curl -LO https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
./install.sh
```
It will create a `bin` directory under your current directory and place `resticprofile.exe` in it.

## Installation using scoop

Resticprofile can be installed from [scoop](https://scoop.sh) main bucket:

```powershell
scoop install resticprofile
```

## Manual installation

- Download the package corresponding to your system and CPU from the [release page](https://github.com/creativeprojects/resticprofile/releases)
- Once downloaded you need to open the archive and copy the binary file `resticprofile` (or `resticprofile.exe`) in your PATH.
