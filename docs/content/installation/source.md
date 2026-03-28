---
title: "From Source"
weight: 13
---

## Installation from source
Requirements: `git`, [Go compiler](https://golang.org/dl/), `make`

Clone the repository by running:
```shell
git clone https://github.com/creativeprojects/resticprofile.git
cd resticprofile
```

To compile the binary, run: `make build`

To install the binary in your user path, run `make install`

To build the binary for all common platforms (`build-mac`, `build-linux`, `build-pi` & `build-windows`), run: `make build-all`

{{% notice style="tip" %}}
Alternatively, a **go-only** build (without `GNU Make`) is accomplished with:
```shell
go build -v -o resticprofile .
```
{{% /notice %}}
