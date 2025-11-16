---
archetype: "home"
title: resticprofile
description: Configuration profiles manager for restic backup
---

Configuration profiles manager for [restic backup](https://restic.net/)

**resticprofile** bridges the gap between a configuration file and restic backup. Although creating a configuration file for restic has been [discussed](https://github.com/restic/restic/issues/16), it remains a low priority task.

The configuration file supports various formats:
* [TOML](https://github.com/toml-lang/toml): files with extensions *.toml* and *.conf* (for compatibility with versions before 0.6.0)
* [JSON](https://en.wikipedia.org/wiki/JSON): files with extension *.json*
* [YAML](https://en.wikipedia.org/wiki/YAML): files with extension *.yaml*
* [HCL](https://github.com/hashicorp/hcl): files with extension *.hcl*
