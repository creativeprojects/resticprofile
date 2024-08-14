# Common mixins library

A library of reusable mixins to simplify common tasks needed when taking backups.

Currently, this library contains mixins for:
* Dumping databases to restic stdin (mysql, mariadb, pgsql)
* Taking temporary snapshots (lvm, btrfs)
* Temporarily freezing VM disk images (virsh/kvm)

## Example Usage

`profiles.toml`

```toml
version = "2"

# downloaded by "https://raw.githubusercontent.com/creativeprojects/resticprofile/master/contrib/mixins/get.sh" 
includes = ["mixins-*.yaml"]

[profiles.__base]
repo = "..."

[proflies.default]
inherit = "__base"
[proflies.default.backup]
use = {name = "snapshot-btrfs", FROM = "/opt/data", TO = "/opt/data_snapshot"}
source = "/opt/data_snapshot"

[proflies.mysql]
inherit = "__base"
[proflies.mysql.backup]
use = {name = "database-mysql", DATABASE = "dbname", USER="dbuser", PASSWORD_FILE="/path/to/password.txt"}

[proflies.vms]
inherit = "__base"
[proflies.vms.backup]
use = [
    {name = "snapshot-virsh", DOMAIN = "vmname1", DUMPXML = "/opt/vms/vmname1.xml"},
    {name = "snapshot-virsh", DOMAIN = "vmname2", DUMPXML = "/opt/vms/vmname2.xml"},
    {name = "snapshot-virsh", DOMAIN = "vmname2", DUMPXML = "/opt/vms/vmname2.xml"},
]
source = "/opt/vms/"
includes = ["*.qcow2", "*.xml"]
```

## Setup

### Posix environment:

```shell
cd /etc/resticprofile \
&& curl -sL https://raw.githubusercontent.com/creativeprojects/resticprofile/master/contrib/mixins/get.sh | sh -
```

### Powershell environment:

At the moment the mixins library doesn't support powershell. Contributions are welcome.

## Disclaimer

Please note that the actions that some of the mixins perform can lead to data loss. This lies in the nature of  
creating and removing snapshots with system privileges. You should carefully read through the actions inside 
the mixins and consider if the mixin is safe for your use case. Please report bugs when you find any.

This library is under the [GPL3](../../LICENSE) license like all other parts of the project.
