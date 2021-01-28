# resticprofile deployment using ansible

This is very much work in progress. Once I get a stable ansible script I should publish it to Ansible Galaxy :)

The playbook is installing:
- latest restic binary to `/usr/local/bin`
- latest resticprofile binary to `/usr/local/bin`
- the resticprofile configuration file from a template file found in `resticprofile/{{ inventory_hostname }}.conf` to `/root/resticprofile/profiles.conf`

## Requirement

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

I might find a way to detect this automatically at some point
