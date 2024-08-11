# Windows CLI: Use a protected folder for system-wide configuration

## Overview

You may wish to do system-wide backups using restic/resticprofile
while ensuring only appropriate users can view the secrets in your
profiles and related files.

This document shows one method of using the command-line to set
up resticprofile for system-wide use with a folder which is only
accessible by the Administrators group and the SYSTEM account.

**NB**: This guide applies to Windows 10 and Windows 11, some
differences may exist with previous versions of Windows.

1. [Overview](#overview)
2. [Procedure](#procedure)
   1. [Create `resticprofile` folder in `ProgramData`](#create-resticprofile-folder-in-programdata)
   2. [Create `resticlogs` folder in `ProgramData`](#create-resticlogs-folder-in-programdata)
   3. [(Optional) Pin `resticlogs` folder to Start and/or 'Quick access\`](#optional-pin-resticlogs-folder-to-start-andor-quick-access)
   4. [Set ACL (permissions) on the `resticprofile` folder](#set-acl-permissions-on-the-resticprofile-folder)
   5. [(Optional) Set ACL (permissions) on the `resticlogs` folder](#optional-set-acl-permissions-on-the-resticlogs-folder)
   6. [Create your resticprofile profiles configuration file](#create-your-resticprofile-profiles-configuration-file)
3. [Final notes](#final-notes)

## Procedure

### Create `resticprofile` folder in `ProgramData`

1. Open a PowerShell Administrative console and execute:

```powershell
C:
cd \ProgramData
mkdir resticprofile
```

### Create `resticlogs` folder in `ProgramData`

From the same console, execute:

```powershell
cd \ProgramData
mkdir resticlogs
```

### (Optional) Pin `resticlogs` folder to Start and/or 'Quick access`

In the same console, issue: `explorer .` to open File Explorer. Then, right-click on the `resticlogs` folder and choose
'Pin to Start' and/or 'Pin to Quick access'.

This along with the optional permissions below will allow you to view
your resticprofile logs without an elevated session.

### Set ACL (permissions) on the `resticprofile` folder

In the same console, execute:

```powershell
icacls resticprofile /inheritance:d
icacls resticprofile /remove:g BUILTIN\Users
```

You should now see (via `icacls .`):

```plaintext
resticprofile NT AUTHORITY\SYSTEM:(OI)(CI)(F)
              BUILTIN\Administrators:(OI)(CI)(F)
              CREATOR OWNER:(OI)(CI)(IO)(F)
```

### (Optional) Set ACL (permissions) on the `resticlogs` folder

In the same console execute:

```powershell
icacls resticlogs /inheritance:d
icacls resticlogs /remove:g BUILTIN\Users
icacls resticlogs /grant "YourDomain/YourUser:(OI)(CI)(RX)"
```

where 'YourDomain' and 'YourUser' are your domain or computer name and the user account
specified above.

### Create your resticprofile profiles configuration file

Using a `toml` profile configuration file for the example

```powershell
New-Item resticprofile\profiles.toml
notepad resticprofile\profiles.toml
```

Edit the configuration and save it. Remember to configure the log files to use
the `C:\\ProgramData\\resticlogs\\` folder.

## Final notes

- Resticprofile can now be used from an Administrative console.

- To view the logs you can use the Start menu or Quick access links you created, or you
  can open an Administrative console and issue:

  ```powershell
  type C:\ProgramData\resticlogs\name-of-log.log
  ```

  for a quick view, or

  ```powershell
  notepad C:\ProgramData\resticlogs\name-of-log.log
  ```

  for more in-depth browsing (especially as the logs get larger).
