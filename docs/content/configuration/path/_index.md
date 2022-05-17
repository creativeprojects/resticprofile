---
title: "Path"
date: 2022-04-24T09:44:41+01:00
weight: 10
---


The default name for the configuration file is `profiles`, without an extension.
You can change the name and its path with the `--config` or `-c` option on the command line.
You can set a specific extension `-c profiles.conf` to load a TOML format file.
If you set a filename with no extension instead, resticprofile will load the first file it finds with any of these extensions:
- .conf (toml format)
- .yaml
- .toml
- .json
- .hcl

## macOS X

resticprofile will search for your configuration file in these folders:
- _current directory_
- ~/Library/Preferences/resticprofile/
- /Library/Preferences/resticprofile/
- /usr/local/etc/
- /usr/local/etc/restic/
- /usr/local/etc/resticprofile/
- /etc/
- /etc/restic/
- /etc/resticprofile/
- /opt/local/etc/
- /opt/local/etc/restic/
- /opt/local/etc/resticprofile/
- ~/ ($HOME directory)

## Other unixes (Linux and BSD)

resticprofile will search for your configuration file in these folders:
- _current directory_
- ~/.config/resticprofile/
- /etc/xdg/resticprofile/
- /usr/local/etc/
- /usr/local/etc/restic/
- /usr/local/etc/resticprofile/
- /etc/
- /etc/restic/
- /etc/resticprofile/
- /opt/local/etc/
- /opt/local/etc/restic/
- /opt/local/etc/resticprofile/
- ~/ ($HOME directory)

## Windows

resticprofile will search for your configuration file in these folders:
- _current directory_
- %USERPROFILE%\AppData\Local\
- c:\ProgramData\
- c:\restic\
- c:\resticprofile\
- %USERPROFILE%\
