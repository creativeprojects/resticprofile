default {
    repository = "rest:http://user:password@localhost:8000/path"
    password-file = "key"
}

simple {
    inherit = "default"
    backup {
        exclude = "/**/.git"
        source = "/source"
    }
}

spaces {
    inherit = "default"
    backup {
        exclude = "My Documents"
        source = "/source dir"
    }
}

quotes {
    inherit = "default"
    backup {
        exclude = ["My'Documents", "My\"Documents"]
        source = ["/source'dir", "/source\"dir"]
    }
}

glob {
    inherit = "default"
    backup {
        exclude = ["examples/integration*"]
        source = ["examples/integration*"]
    }
}

mixed {
    inherit = "default"
    backup {
        exclude = ["examples/integration*"]
        source = ["/Côte d'Ivoire"]
    }
}