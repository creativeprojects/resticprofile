
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
* You can run [shell commands]({{< ref "/configuration/run_hooks" >}}) before or after running a profile: useful if you need to mount and unmount your backup disk for example
* You can run a [shell command]({{< ref "/configuration/run_hooks" >}}) if an error occurred (at any time)
* You can send a backup stream via _stdin_
* You can start restic at a lower or higher priority (Priority Class in Windows, *nice* in all unixes) and/or _ionice_ (only available on Linux)
* It can check that you have [enough memory]({{< ref "/usage/memory" >}}) before starting a backup. (I've had some backups that literally killed a server with swap disabled)
* You can generate cryptographically secure random keys to use as a restic [key file]({{< ref "/usage/keyfile" >}})
* You can easily [schedule]({{< ref "/schedules" >}}) backups, retentions and checks (works for *systemd*, *crond*, *launchd* and *windows task scheduler*)
* You can generate a simple [status file]({{< ref "/status" >}}) to send to some monitoring software and make sure your backups are running fine 
* You can use a template syntax in your configuration file
* You can generate scheduled tasks using *crond*
* Get backup statistics in your [status file]({{< ref "/status" >}})
* Automatically clear up [stale locks]({{< ref "/usage/locks" >}})
* Export a [prometheus]({{< ref "/status/prometheus" >}}) file after a backup, or send the report to a push gateway automatically
* **[new for v0.16.0]** Full support for the [copy]({{< ref "/configuration/copy" >}}) command (with scheduling)
* **[new for v0.16.0]** Describe your own [systemd units and timers]({{< ref "/schedules/systemd" >}}) with go templates
* **[new for v0.17.0]** Run shell commands in the background when non fatal errors are detected from restic
* **[new for v0.18.0]** Send messages to [HTTP hooks]({{< ref "/configuration/http_hooks" >}}) before, after a successful or failed job (backup, forget, check, prune, copy)
* **[new for v0.18.0]** Automatically initialize the secondary repository using `copy-chunker-params` flag
* **[new for v0.18.0]** Send resticprofile logs to a syslog server

The configuration file accepts various formats:
* [TOML](https://github.com/toml-lang/toml) : configuration file with extension _.toml_ and _.conf_ to keep compatibility with versions before 0.6.0
* [JSON](https://en.wikipedia.org/wiki/JSON) : configuration file with extension _.json_
* [YAML](https://en.wikipedia.org/wiki/YAML) : configuration file with extension _.yaml_
* [HCL](https://github.com/hashicorp/hcl): configuration file with extension _.hcl_

