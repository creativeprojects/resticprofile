
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

global_flags = {
    'ionice': { 'type': 'bool' },
    'ionice-class': { 'type': 'int' },
    'ionice-level': { 'type': 'int' },
    'nice': { 'type': [ 'bool', 'int' ] },
    'default-command': { 'type': 'str' },
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
        'repository': { 'type': 'str', 'flag': 'repo' },
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

def validate_configuration_option(definition, key, value):
    if key in definition:
        if 'type' in definition[key]:
            if isinstance(definition[key]['type'], list):
                for expected_type in definition[key]['type']:
                    success = check_type(expected_type, value, expect_list=('list' in definition[key] and definition[key]['list']))
                    if success:
                        return {'key': key, 'value': value, 'type': expected_type}
                return False
            else:
                expected_type = definition[key]['type']
                success = check_type(expected_type, value, expect_list=('list' in definition[key] and definition[key]['list']))
                if success:
                    return {'key': key, 'value': value, 'type': expected_type}
                return False
        else:
            return False
    else:
        return False

def check_type(expected_type, value, expect_list = False):
    if expect_list and isinstance(value, list):
        for subvalue in value:
            success = check_type(expected_type, subvalue, expect_list = False)
            if not success:
                return False
        return True

    if expected_type == 'bool':
        return isinstance(value, bool)
    elif expected_type == 'int':
        return isinstance(value, int)
    elif expected_type in ('str', 'file'):
        return isinstance(value, str)
    else:
        raise Exception("Unknown type '{}'".format(expected_type))
