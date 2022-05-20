default {
    repository = "local:/backup"
    password-file = "key"

    backup = {
        verbose = true
        source = [ "/home" ]
    }
}