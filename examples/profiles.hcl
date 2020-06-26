global = {
    default-command = "version"
    initialize = false
    priority = "low"
}

default = {
    repository = "/Volumes/RAMDisk"
    password-file = "key"
    env = {
        TMPDIR = "/tmp"
    }
}


groups = {
    full-backup = ["root", "src"]
}

linux = {
    repository =  "/tmp/restic"
    password-file =  "key"

    env =  {
        TMPDIR = "/tmp"
    }

    backup =  {
        "source" = ["."]
        "tag" = ["linux", "dev"]
    }

    retention =  {
        "after-backup" = true
        "before-backup" = false
        "compact" = false
        "keep-daily" = 1
        "keep-hourly" = 1
        "keep-last" = 3
        "keep-monthly" = 1
        "keep-tag" = ["forever"]
        "keep-weekly" = 1
        "keep-within" = "3h"
        "keep-yearly" = 1
        "prune" = false
    }

    snapshots = {
        initialize = true
        tag = ["linux", "dev"]
    }
}

no-cache = {
    inherit = "default"
    no-cache = true
}

root = {
    inherit = "default"
    initialize = true

    backup =  {
        "exclude-caches" = true
        "exclude-file" = ["root-excludes", "excludes"]
        "one-file-system" = false
        "source" = ["."]
        "tag" = ["test", "dev"]
    }

    retention =  {
        "after-backup" = true
        "before-backup" = false
        "compact" = false
        "host" = true
        "keep-daily" = 1
        "keep-hourly" = 1
        "keep-last" = 3
        "keep-monthly" = 1
        "keep-tag" = ["forever"]
        "keep-weekly" = 1
        "keep-within" = "3h"
        "keep-yearly" = 1
        "prune" = false
        "tag" = ["test", "dev"]
    }
}

src = {
    inherit = "default"
    initialize = true
    "run-before" = "echo mount backup disk"
    "run-after" = ["echo sync", "echo umount backup disk"]

    backup =  {
        "check-before" = true
        "exclude" = ["/**/.git"]
        "exclude-caches" = true
        "one-file-system" = false
        "run-after" = "echo All Done!"
        "run-before" = ["echo Starting!", "ls -al ./src"]
        "source" = ["./src"]
        "tag" = ["test", "dev"]
    }

    retention =  {
        "after-backup" = true
        "before-backup" = false
        "compact" = false
        "keep-within" = "30d"
        "prune" = true
    }

}

stdin = {
    inherit = "default"

    backup =  {
        stdin = true
        stdin-filename = "stdin-test"
        tag = ["stdin"]
    }
}