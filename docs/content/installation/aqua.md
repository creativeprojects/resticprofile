---
title: "Aqua / Mise"
weight: 55
---

## Installation from aqua or mise

This is assuming you're already using [aqua](https://aquaproj.github.io) or [mise](https://mise.jdx.dev): there's no need to install these package managers just to install resticprofile 😊

resticprofile is available in the standard aqua registry starting on version `v4.586.0`

### using aqua CLI

```shell
aqua generate -i creativeprojects/resticprofile
aqua install
```

### aqua.yaml

<!-- checkdoc-ignore -->
```yaml
registries:
- type: standard
  ref: v4.586.0 # renovate: depName=aquaproj/aqua-registry
packages:
- name: creativeprojects/resticprofile@v0.34.0

```
