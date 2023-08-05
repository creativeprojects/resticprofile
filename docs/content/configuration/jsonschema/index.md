---
title: "JSON schema"
date: 2022-04-24T09:44:47+01:00
tags: ["v0.21.0"]
weight: 55
---

JSON, YAML and TOML configuration files can benefit from a JSON schema that describes the 
config structure depending on the selected *restic* and configuration file version.

## Schema URL

**{{< absolute "jsonschema/config.json" >}}**

This schema detects config and restic version and redirects to the corresponding [versioned URL](#versioned-urls).
Detection requires full support for JSON Schema Draft 7. Use a versioned URL where not supported (e.g. TOML). 

## Usage (vscode)

To use a schema with **vscode**, begin your config files with:

{{< tabs groupId="config-with-json" >}}
{{% tab name="toml" %}}
``````toml
#:schema {{< absolute "jsonschema/config-2.json" nohtml >}}

version = "2"
 
``````
{{% /tab %}}
{{% tab name="yaml" %}}
``````yaml
# yaml-language-server: $schema={{< absolute "jsonschema/config.json" nohtml >}}

version: "2"
 
``````
{{% /tab %}}
{{% tab name="json" %}}
``````json
{
    "$schema": "{{< absolute "jsonschema/config.json" nohtml >}}",
    "version": "2"
}
``````
{{% /tab %}}
{{% /tabs %}}

{{% notice style="hint" %}}
YAML & TOML validation with JSON schema is not supported out of the box. Additional extensions are required.
{{% /notice %}}

**Example**

{{< figure src="/recordings/jsonschema-vsc.gif" >}}

**Extension**: `redhat.vscode-yaml`

## Versioned URLs

The following versioned schemas are referenced from `config.json`. Which URL is selected depends 
on config and restic version. You may use the URLs directly if you need a self-contained schema 
file or to enforce a certain version of configuration format and/or restic flags

JSON schema URLs for **any** *restic* version:

* Config V1: {{< absolute "jsonschema/config-1.json" >}}
* Config V2: {{< absolute "jsonschema/config-2.json" >}}

These universal schemas contain all flags and commands of all known *restic* versions and 
may allow to set flags that are not supported by a particular *restic* version. Descriptions 
and deprecation markers indicate what is supported and what should no longer be used.

JSON schema URLs for a **specific** *restic* version (list of [available URLs]({{< ref "reference/#json-schema" >}})):

* `.../config-1-restic-{MAJOR}-{MINOR}.json`
* `.../config-2-restic-{MAJOR}-{MINOR}.json`

These schemas contain only flags and commands of a specific *restic* version. The schema will 
validate a config only when flags are used that the selected *restic* version supports 
according to its manual pages.

{{% notice style="hint" %}}
As an alternative to URLs, schemas can be generated locally with: 
`resticprofile generate --json-schema [--version RESTIC_VERSION] [v2|v1]`
{{% /notice %}}
