---
title: "Snapshot filters"
weight: 12
---

Several restic commands filter snapshots with `host`, `path` and `tag`.
In resticprofile these options also accept a boolean:

* `true` — copy the value from the `backup` section (or the system hostname for `host`)
* `false` — do not pass the filter to restic
* a string / list — use that value as usual

The **default when the option is omitted** depends on the section and on the
[configuration version]({{% relref "/configuration/v2" %}}).
The generated [configuration reference]({{% relref "/reference" %}}) leaves the
**Default** column empty for these options because a single static default would
be misleading across versions; the behaviour is summarized below.

{{% notice style="note" %}}
Automatic defaults apply to the **`retention`** section (and to `backup.host` in
configuration version 2). The standalone **`forget`** section — and other
commands such as `snapshots` or `check` — do **not** set `host` / `path` / `tag`
automatically; set them explicitly (including `true` to copy from `backup`).
{{% /notice %}}

## `retention` section

`retention` runs `restic forget` before or after `backup`. When a filter is left
unset, resticprofile may inject `true` so that forget selects the same snapshots
the profile backs up.

| Option | Configuration version 1 | Configuration version 2 |
|--------|-------------------------|-------------------------|
| `path` | Defaults to `true` when a `backup` section exists | Same as version 1 |
| `tag` | Unset (tags are not used as a filter) | Defaults to `true` when `backup.tag` is set |
| `host` | Unset | Defaults to `true` (system hostname), or copies `backup.host` when that is set |

Evidence in source: `RetentionSection.resolve` in `config/profile.go` always sets
`path` to `true` when it is missing and `backup` is present; the `tag` and `host`
defaults are gated on `version >= 2`.

{{% notice style="tip" %}}
If you change `backup.source` while `path` stays at its default `true`, retention
and other path-filtered commands will target the **new** paths and may no longer
match older snapshots. Prefer an explicit `path`, or a tag-centric filter
(`tag: true` with a stable `backup.tag`), when sources move.
{{% /notice %}}

## `backup` section

| Option | Configuration version 1 | Configuration version 2 |
|--------|-------------------------|-------------------------|
| `host` | Unset (restic uses `$RESTIC_HOST` / the system hostname) | Defaults to `true` (system hostname) |
| `path` / `tag` | Boolean `true` is unsupported | Same |

## Explicit `true` / `false`

When you set the option yourself, resolution is the same in both versions:

* `host: true` → replaced with the system hostname (`retention` uses the hostname
  that applies in `backup` when that section sets `host`)
* `path: true` → replaced with the resolved `backup.source` paths
* `tag: true` → replaced with `backup.tag` (if `backup` has no tags, no `--tag`
  flag is added)
* `false` → the filter is omitted

## Related reference notes

The generated reference already mentions these defaults in the **Notes** column
for `retention` (and for `backup.host` in version 2). This page is the
version-aware overview requested for issue
[#200](https://github.com/creativeprojects/resticprofile/issues/200).