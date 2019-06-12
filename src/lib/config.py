
defaults = {
    'config_file': 'profiles.conf',
    'profile': 'default',
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
