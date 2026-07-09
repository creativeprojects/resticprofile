---
title: Profiles
weight: 30
alwaysopen: false
---

## Choose Your Favorite Format

The **resticprofile** configuration file can be written in:
* [TOML](https://github.com/toml-lang/toml): configuration file with extension *.toml* or *.conf*
* [YAML](https://en.wikipedia.org/wiki/YAML): configuration file with extension *.yaml*
* [JSON](https://en.wikipedia.org/wiki/JSON): configuration file with extension *.json*
* [HCL](https://github.com/hashicorp/hcl): configuration file with extension *.hcl*

We recommend using either TOML or YAML.

JSON is suitable for auto-generated configurations but is not the easiest format for humans to read and write.

HCL can be useful if you already use a tool from the Hashicorp stack; otherwise, it's another format to learn.

## Debugging your template and variable expansion

If for some reason you don't understand why resticprofile is not loading your configuration file, you can display the generated configuration after executing the template (and replacing the variables and everything) using the `--trace` flag.

## More Information

{{% children %}}
