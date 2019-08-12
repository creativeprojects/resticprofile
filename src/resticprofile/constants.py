'''
All resticprofile constants here
'''

DEFAULT_CONFIGURATION_FILE = 'profiles.conf'
DEFAULT_PROFILE_NAME = 'default'
DEFAULT_COMMAND = 'snapshots'
DEFAULT_IONICE_FLAG = False
DEFAULT_INITIALIZE_FLAG = False
DEFAULT_VERBOSE_FLAG = None
DEFAULT_QUIET_FLAG = None

ARGUMENTS_DEFINITION = {
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
    },
    'no-ansi': {
        'long': 'no-ansi',
        'argument': False,
    }
}

DEFINITION_TYPE = 'type'
DEFINITION_FLAG = 'flag'

SECTION_CONFIGURATION_GLOBAL = 'global'
SECTION_CONFIGURATION_RETENTION = 'retention'
SECTION_CONFIGURATION_ENVIRONMENT = 'env'
SECTION_CONFIGURATION_GROUPS = 'groups'

SECTION_DEFINITION_COMMON = 'common'
SECTION_DEFINITION_FORGET = 'forget'

PARAMETER_IONICE = 'ionice'
PARAMETER_IONICE_CLASS = 'ionice-class'
PARAMETER_IONICE_LEVEL = 'ionice-level'

PARAMETER_NICE = 'nice'
PARAMETER_DEFAULT_COMMAND = 'default-command'
PARAMETER_INITIALIZE = 'initialize'
PARAMETER_INHERIT = 'inherit'
PARAMETER_REPOSITORY = 'repository'
PARAMETER_REPO = 'repo'
PARAMETER_QUIET = 'quiet'
PARAMETER_VERBOSE = 'verbose'
PARAMETER_RESTIC_BINARY = 'restic-binary'
PARAMETER_SOURCE = 'source'
PARAMETER_FORGET_BEFORE_BACKUP = 'before-backup'
PARAMETER_FORGET_AFTER_BACKUP = 'after-backup'
PARAMETER_PATH = 'path'
PARAMETER_HOST = 'host'
PARAMETER_CHECK_BEFORE = 'check-before'
PARAMETER_CHECK_AFTER = 'check-after'
PARAMETER_RUN_BEFORE = 'run-before'
PARAMETER_RUN_AFTER = 'run-after'
PARAMETER_STDIN = 'stdin'

COMMAND_BACKUP = 'backup'
COMMAND_CHECK = 'check'
COMMAND_FORGET = 'forget'
COMMAND_INIT = 'init'
COMMAND_PRUNE = 'prune'
COMMAND_SNAPSHOTS = 'snapshots'
