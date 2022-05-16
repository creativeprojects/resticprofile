![Build](https://github.com/creativeprojects/resticprofile/workflows/Build/badge.svg)
[![codecov](https://codecov.io/gh/creativeprojects/resticprofile/branch/master/graph/badge.svg?token=cUozgF9j4I)](https://codecov.io/gh/creativeprojects/resticprofile)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativeprojects/resticprofile)](https://goreportcard.com/report/github.com/creativeprojects/resticprofile)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=ncloc)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=bugs)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=vulnerabilities)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=sqale_rating)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=reliability_rating)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=creativeprojects_resticprofile&metric=security_rating)](https://sonarcloud.io/dashboard?id=creativeprojects_resticprofile)

# resticprofile
Configuration profiles manager for [restic backup](https://restic.net/)

**resticprofile** is the missing link between a configuration file and restic backup. Creating a configuration file for restic has been [discussed before](https://github.com/restic/restic/issues/16), but seems to be a very low priority right now.

With resticprofile:

* You no longer need to remember command parameters and environment variables
* You can create multiple profiles inside one configuration file
* A profile can inherit all the options from another profile
* You can run the forget command before or after a backup (in a section called *retention*)
* You can check a repository before or after a backup
* You can create groups of profiles that will run sequentially
* You can run shell commands before or after running a profile: useful if you need to mount and unmount your backup disk for example
* You can run a shell command if an error occurred (at any time)
* You can send a backup stream via _stdin_
* You can start restic at a lower or higher priority (Priority Class in Windows, *nice* in all unixes) and/or _ionice_ (only available on Linux)
* It can check that you have enough memory before starting a backup. (I've had some backups that literally killed a server with swap disabled)
* You can generate cryptographically secure random keys to use as a restic key file
* You can easily schedule backups, retentions and checks (works for *systemd*, *crond*, *launchd* and *windows task scheduler*)
* You can generate a simple status file to send to some monitoring software and make sure your backups are running fine 
* You can use a template syntax in your configuration file
* You can generate scheduled tasks using *crond*
* Get backup statistics in your status file
* **[new for v0.14.0]** Automatically clear up [stale locks](#locks)
* **[new for v0.15.0]** Export a **prometheus** file after a backup, or send the report to a push gateway automatically
* **[new for v0.16.0]** Full support for the copy command (with scheduling)
* **[new for v0.16.0]** Describe your own systemd units and timers with go templates

The configuration file accepts various formats:
* [TOML](https://github.com/toml-lang/toml) : configuration file with extension _.toml_ and _.conf_ to keep compatibility with versions before 0.6.0
* [JSON](https://en.wikipedia.org/wiki/JSON) : configuration file with extension _.json_
* [YAML](https://en.wikipedia.org/wiki/YAML) : configuration file with extension _.yaml_
* [HCL](https://github.com/hashicorp/hcl): configuration file with extension _.hcl_

For the rest of the documentation, I'll be showing examples using different formats, but mostly TOML and YAML.

# Table of Contents

This README is getting too big now, so I'm in the process of moving most of its content into a proper documentation.

**This is work in progress** but you can still [check the documentation out here](https://creativeprojects.github.io/resticprofile/).

<!--ts-->
* [resticprofile](#resticprofile)
* [Table of Contents](#table-of-contents)
* [Requirements](#requirements)
* [Installation](#installation)
  * [Homebrew (all macOS, Linux on amd64)](#homebrew-all-macos-linux-on-amd64)
    * [Note on installing on Linux via Homebrew](#note-on-installing-on-linux-via-homebrew)
  * [Via a script (macOS, Linux &amp; other unixes)](#via-a-script-macos-linux--other-unixes)
  * [Installation for Windows using bash](#installation-for-windows-using-bash)
  * [Manual installation (Windows)](#manual-installation-windows)
  * [Ansible](#ansible)
  * [Installation from source](#installation-from-source)
  * [Shell completion](#shell-completion)
* [Upgrade](#upgrade)
* [Using docker image](#using-docker-image)
  * [Container host name](#container-host-name)
* [Configuration format](#configuration-format)
* [Configuration examples](#configuration-examples)
  * [Simple TOML configuration](#simple-toml-configuration)
  * [Simple YAML configuration](#simple-yaml-configuration)
  * [More complex configuration in TOML](#more-complex-configuration-in-toml)
  * [TOML configuration example for Windows](#toml-configuration-example-for-windows)
  * [Use stdin in configuration](#use-stdin-in-configuration)
  * [Special case for the copy command section](#special-case-for-the-copy-command-section)
* [Configuration paths](#configuration-paths)
  * [macOS X](#macos-x)
  * [Other unixes (Linux and BSD)](#other-unixes-linux-and-bsd)
  * [Windows](#windows)
* [Includes](#includes)
* [Path resolution in configuration](#path-resolution-in-configuration)
* [Run commands before, after success or after failure](#run-commands-before-after-success-or-after-failure)
  * [order of run\-\* during a backup](#order-of-run--during-a-backup)
* [Warnings from restic](#warnings-from-restic)
  * [no\-error\-on\-warning](#no-error-on-warning)
* [Locks](#locks)
  * [Stale locks](#stale-locks)
  * [Lock wait](#lock-wait)
  * [restic lock management](#restic-lock-management)
* [Using resticprofile](#using-resticprofile)
* [Command line reference](#command-line-reference)
* [Minimum memory required](#minimum-memory-required)
* [Version](#version)
* [Generating random keys](#generating-random-keys)
* [Scheduled backups](#scheduled-backups)
  * [retention schedule is deprecated](#retention-schedule-is-deprecated)
  * [Schedule configuration](#schedule-configuration)
    * [schedule\-permission](#schedule-permission)
    * [schedule\-lock\-mode](#schedule-lock-mode)
    * [schedule\-lock\-wait](#schedule-lock-wait)
    * [schedule\-log](#schedule-log)
    * [schedule\-priority (systemd and launchd only)](#schedule-priority-systemd-and-launchd-only)
    * [schedule](#schedule)
  * [Scheduling commands](#scheduling-commands)
    * [schedule command](#schedule-command)
    * [unschedule command](#unschedule-command)
    * [status command](#status-command)
    * [Examples of scheduling commands under Windows](#examples-of-scheduling-commands-under-windows)
    * [Examples of scheduling commands under Linux](#examples-of-scheduling-commands-under-linux)
    * [Examples of scheduling commands under macOS](#examples-of-scheduling-commands-under-macos)
  * [Changing schedule\-permission from user to system, or system to user](#changing-schedule-permission-from-user-to-system-or-system-to-user)
* [Status file for easy monitoring](#status-file-for-easy-monitoring)
  * [Extended status](#extended-status)
* [Prometheus](#prometheus)
  * [User defined labels](#user-defined-labels)
* [Variable expansion in configuration file](#variable-expansion-in-configuration-file)
  * [Pre\-defined variables](#pre-defined-variables)
  * [Hand\-made variables](#hand-made-variables)
  * [Examples](#examples)
* [Configuration templates](#configuration-templates)
* [Debugging your template and variable expansion](#debugging-your-template-and-variable-expansion)
* [Limitations of using templates](#limitations-of-using-templates)
* [Documentation on template, variable expansion and other configuration scripting](#documentation-on-template-variable-expansion-and-other-configuration-scripting)
* [Configuration file reference](#configuration-file-reference)
* [Appendix](#appendix)
* [Using resticprofile and systemd](#using-resticprofile-and-systemd)
  * [systemd calendars](#systemd-calendars)
  * [First time schedule](#first-time-schedule)
  * [How to change the default systemd unit and timer file using a template](#how-to-change-the-default-systemd-unit-and-timer-file-using-a-template)
    * [Default unit file](#default-unit-file)
    * [Default timer file](#default-timer-file)
    * [Template variables](#template-variables)
* [Using resticprofile and launchd on macOS](#using-resticprofile-and-launchd-on-macos)
  * [User agent](#user-agent)
    * [Special case of schedule\-permission=user with sudo](#special-case-of-schedule-permissionuser-with-sudo)
  * [Daemon](#daemon)
* [Contributions](#contributions)

<!--te-->



