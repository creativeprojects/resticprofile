---
title: "Ansible"
date: 2022-04-23T23:47:41+01:00
weight: 15
---

## Installation using Ansible

Installation using Ansible is not streamlined, but here's the playbook I'm using on my servers:

{{%attachments title="Playbooks" pattern=".*"/%}}

This is very much work in progress. Once I get a stable ansible script I should publish it to Ansible Galaxy.

The playbook is installing (or upgrading):

* latest restic binary to `/usr/local/bin`
* latest resticprofile binary to `/usr/local/bin`
* the resticprofile configuration file from a template file found in `/resticprofile/{{ inventory_hostname }}/profiles.conf` to `/root/resticprofile/profiles.conf`
* other files (like files needed for `--exclude-file`, `--files-from` or anything else you need) from `/resticprofile/{{ inventory_hostname }}/copy/*` to `/root/resticprofile/`

### Requirement

Each target machine must have one variable `arch` containing the resticprofile OS & Arch. You can see a list on a download page.

Typically, a binary will be distributed using this convention:

`resticprofile-[VERSION]_[OS]_[ARCH].tar.gz`

Your host variables file should declare a `arch` variable containing the `[OS]_[ARCH]` part of the file name:

```
arch: linux_amd64
```

or for a Raspberry pi 3+:

```
arch: linux_armv7
```

Note: _I might find a way to detect this automatically at some point_
