global {
    priority = "low"
    ionice = true
    ionice-class = 2
    ionice-level = 6
    # don't start if the memory available is < 1000MB
    min-memory = 1000
}

groups {
    all = ["src", "self"]
}

default {
    repository = "/tmp/backup"
    password-file = "key"
    run-before = "echo Profile started!"
    run-after = "echo Profile finished!"
    run-after-fail = "echo An error occured!"
}


src {
    inherit = "default"
    initialize = true
    lock = "/tmp/backup/resticprofile-profile-src.lock"

    snapshots = {
        tag = [ "test", "dev" ]
    }

    backup = {
        run-before = [ "echo Starting!", "ls -al ~/go/src" ]
        run-after = "echo All Done!"
        exclude = [ "/**/.git" ]
        exclude-caches = true
        tag = [ "test", "dev" ]
        source = [ "~/go/src" ]
        check-before = true
    }

    retention = {
        before-backup = false
        after-backup = true
        keep-last = 3
        compact = false
        prune = true
    }

    check = {
        check-unused = true
        with-cache = false
    }
}
