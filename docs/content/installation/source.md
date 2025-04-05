---
title: "Source"
weight: 16
---

## Installation from source

It's very easy to compile from the source code.

Ensure your machine has the following:  
- `git` (use `git-bash` on Windows)  
- [Go compiler](https://golang.org/dl/)  
- `GNU Make` (preinstalled on many Unix systems). On Debian-based distributions (e.g., Ubuntu), install the `build-essential` package.  

Compilation:
```shell
git clone https://github.com/creativeprojects/resticprofile.git
cd resticprofile
make build
```

Your compiled binary (`resticprofile` or `resticprofile.exe`) is available in the current folder.

To install the binary in your user path:

```shell
make install
```

To build all common platforms (`build-mac`, `build-linux`, `build-pi` & `build-windows`):

```shell
make build-all
```

Alternatively, a **go-only** build (without `GNU Make`) is accomplished with:

```shell
git clone https://github.com/creativeprojects/resticprofile.git
cd resticprofile
go build -v -o resticprofile .
```
