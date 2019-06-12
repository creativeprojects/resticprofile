
defaults = {
    'configuration_file': 'profiles.conf',
    'profile_name': 'default',
    'global': 'global',
    'separator': '.',
    'environment': 'env',
    'default-command': 'snapshots',
    'ionice': False,
    'initialize': False,
    'verbose': None,
    'quiet': None,
}

arguments_definition = {
    'help': {
        'short': 'h',
        'long': 'help',
        'argument': False,
    },
    'quiet': {
        'short': 'q',
        'long': 'quiet',
        'argument': False,
    },
    'verbose': {
        'short': 'v',
        'long': 'verbose',
        'argument': False,
    },
    'config': {
        'short': 'c',
        'long': 'config',
        'argument': True,
        'argument_name': 'configuration_file'
    },
    'name': {
        'short': 'n',
        'long': 'name',
        'argument': True,
        'argument_name': 'profile_name'
    }
}

restic_flags = {
    'global': {
        'cacert': { 'type': 'file' },
        'cache-dir': { 'type': 'str' },
        'cleanup-cache': { 'type': 'bool' },
        'json': { 'type': 'bool' },
        'key-hint': { 'type': 'str' },
        'limit-download': { 'type': 'int' },
        'limit-upload': { 'type': 'int' },
        'no-cache': { 'type': 'bool' },
        'no-lock': { 'type': 'bool' },
        'password-command': { 'type': 'str' },
        'password-file': { 'type': 'str' },
        'quiet': { 'type': 'bool' },
        'repository': { 'type': 'str', 'list': True },
        'tls-client-cert': { 'type': 'str' },
        'verbose': { 'type': ['bool', 'int'] },
    },
    'backup': {
        'exclude': { 'type': 'str', 'list': True },
        'exclude-caches': { 'type': 'bool' },
        'exclude-file file': { 'type': 'file', 'list': True },
        'exclude-if-present': { 'type': 'str', 'list': True },
        'files-from': { 'type': 'str', 'list': True },
        'force': { 'type': 'bool' },
        'host': { 'type': 'str' },
        'iexclude': { 'type': 'str', 'list': True },
        'ignore-inode': { 'type': 'bool' },
        'one-file-system': { 'type': 'bool' },
        'parent': { 'type': 'str' },
        'stdin': { 'type': 'bool' },
        'stdin-filename': { 'type': 'file' },
        'tag': { 'type': 'str', 'list': True },
        'time': { 'type': 'str' },
        'with-atime': { 'type': 'bool' },
    },
    'snapshots': {
        'compact': { 'type': 'bool' },
        'group-by': { 'type': 'str' },
        'host': { 'type': 'str' },
        'last': { 'type': 'bool' },
        'path': { 'type': 'str', 'list': True },
        'tag': { 'type': 'str', 'list': True },
    },
    "forget": {
        'keep-last': { 'type': 'int' },
        'keep-hourly': { 'type': 'int' },
        'keep-daily': { 'type': 'int' },
        'keep-weekly': { 'type': 'int' },
        'keep-monthly': { 'type': 'int' },
        'keep-yearly': { 'type': 'int' },
        'keep-within': { 'type': 'int' },
        'keep-tag': { 'type': 'str', 'list': True },
        'host': { 'type': 'str' },
        'tag': { 'type': 'str', 'list': True },
        'path': { 'type': 'str', 'list': True },
        'compact': { 'type': 'bool' },
        'group-by': { 'type': 'str' },
        'dry-run': { 'type': 'bool' },
        'prune': { 'type': 'bool' },
    },
    "check": {
        'check-unused': { 'type': 'bool' },
        'read-data': { 'type': 'bool' },
        'read-data-subset': { 'type': 'str' },
        'with-cache': { 'type': 'bool' },
    },
}