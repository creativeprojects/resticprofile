'''
resticprofile configuration
'''
from typing import Union, List
from . import constants
from .flag import Flag
from .ionice import IONice
from .nice import Nice

DEFAULTS = {
    'configuration_file': 'profiles.conf',
    'profile_name': 'default',
    'global': 'global',
    'separator': '.',
    'environment': 'env',
    'default_command': 'snapshots',
    'ionice': False,
    'initialize': False,
    'verbose': None,
    'quiet': None,
}

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
    }
}

GLOBAL_FLAGS_DEFINITION = {
    'ionice': {'type': 'bool'},
    'ionice-class': {'type': 'int'},
    'ionice-level': {'type': 'int'},
    'nice': {'type': ['bool', 'int']},
    'default-command': {'type': 'str'},
    'initialize': {'type': 'bool'},
}

CONFIGURATION_FLAGS_DEFINITION = {
    'common': {
        'inherit': {'type': 'str'},
        'cacert': {'type': 'file'},
        'cache-dir': {'type': 'str'},
        'cleanup-cache': {'type': 'bool'},
        'json': {'type': 'bool'},
        'key-hint': {'type': 'str'},
        'limit-download': {'type': 'int'},
        'limit-upload': {'type': 'int'},
        'no-cache': {'type': 'bool'},
        'no-lock': {'type': 'bool'},
        'option': {'type': 'str', 'list': True},
        'password-command': {'type': 'str'},
        'password-file': {'type': 'file'},
        'quiet': {'type': 'bool'},
        'repository': {'type': 'str', 'flag': 'repo'},
        'tls-client-cert': {'type': 'file'},
        'verbose': {'type': ['bool', 'int']},
    },
    'backup': {
        'exclude': {'type': 'str', 'list': True},
        'exclude-caches': {'type': 'bool'},
        'exclude-file': {'type': 'file', 'list': True},
        'exclude-if-present': {'type': 'str', 'list': True},
        'files-from': {'type': 'str', 'list': True},
        'force': {'type': 'bool'},
        'host': {'type': 'str'},
        'iexclude': {'type': 'str', 'list': True},
        'ignore-inode': {'type': 'bool'},
        'one-file-system': {'type': 'bool'},
        'parent': {'type': 'str'},
        'stdin': {'type': 'bool'},
        'stdin-filename': {'type': 'file'},
        'tag': {'type': 'str', 'list': True},
        'time': {'type': 'str'},
        'with-atime': {'type': 'bool'},
    },
    'snapshots': {
        'compact': {'type': 'bool'},
        'group-by': {'type': 'str'},
        'host': {'type': 'str'},
        'last': {'type': 'bool'},
        'path': {'type': 'str', 'list': True},
        'tag': {'type': 'str', 'list': True},
    },
    "forget": {
        'keep-last': {'type': 'int'},
        'keep-hourly': {'type': 'int'},
        'keep-daily': {'type': 'int'},
        'keep-weekly': {'type': 'int'},
        'keep-monthly': {'type': 'int'},
        'keep-yearly': {'type': 'int'},
        'keep-within': {'type': 'int'},
        'keep-tag': {'type': 'str', 'list': True},
        'host': {'type': 'str'},
        'tag': {'type': 'str', 'list': True},
        'path': {'type': 'str', 'list': True},
        'compact': {'type': 'bool'},
        'group-by': {'type': 'str'},
        'dry-run': {'type': 'bool'},
        'prune': {'type': 'bool'},
    },
    "check": {
        'check-unused': {'type': 'bool'},
        'read-data': {'type': 'bool'},
        'read-data-subset': {'type': 'str'},
        'with-cache': {'type': 'bool'},
    },
}


class Config:
    '''
    Manage configuration information from configuration dictionnary
    '''

    def __init__(self, configuration: dict):
        self.configuration = configuration

    def get_ionice(self, section="global") -> IONice:
        if section not in self.configuration:
            return None

        configuration = self.configuration[section]
        if constants.PARAMETER_IONICE not in configuration:
            return None

        ionice = self.validate_global_configuration_option(
            constants.PARAMETER_IONICE,
            configuration[constants.PARAMETER_IONICE]
        )
        option = None
        if ionice and ionice.value:
            io_class = None
            io_level = None
            if constants.PARAMETER_IONICE_CLASS in configuration:
                io_class = self.validate_global_configuration_option(
                    constants.PARAMETER_IONICE_CLASS,
                    configuration[constants.PARAMETER_IONICE_CLASS]
                )
            if constants.PARAMETER_IONICE_LEVEL in configuration:
                io_level = self.validate_global_configuration_option(
                    constants.PARAMETER_IONICE_LEVEL,
                    configuration[constants.PARAMETER_IONICE_LEVEL]
                )

            if io_class and io_level:
                option = IONice(io_class.value, io_level.value)
            elif io_class:
                option = IONice(io_class=io_class.value)
            elif io_level:
                option = IONice(io_level=io_level.value)
            else:
                option = IONice()

        return option

    def get_nice(self, section="global") -> Nice:
        if section not in self.configuration:
            return None

        configuration = self.configuration[section]
        option = None
        if constants.PARAMETER_NICE in configuration:
            nice = self.validate_global_configuration_option(
                constants.PARAMETER_NICE,
                configuration[constants.PARAMETER_NICE]
            )
            if nice and (nice.value and nice.type == 'bool') or (nice.type == 'int'):
                if nice.type == 'int':
                    option = Nice(nice.value)
                else:
                    option = Nice()

        return option

    def get_default_command(self, section="global") -> str:
        if section not in self.configuration:
            return DEFAULTS['default_command']

        configuration = self.configuration[section]
        if constants.PARAMETER_DEFAULT_COMMAND in configuration:
            default_command = self.validate_global_configuration_option(
                constants.PARAMETER_DEFAULT_COMMAND,
                configuration[constants.PARAMETER_DEFAULT_COMMAND]
            )
            if default_command:
                return default_command.value

        return DEFAULTS['default_command']

    def get_initialize(self, section="global") -> bool:
        if section not in self.configuration:
            return DEFAULTS['initialize']

        configuration = self.configuration[section]
        if constants.PARAMETER_INITIALIZE in configuration:
            initialize = self.validate_global_configuration_option(
                constants.PARAMETER_INITIALIZE,
                configuration[constants.PARAMETER_INITIALIZE]
            )
            if initialize:
                return initialize.value

        return DEFAULTS['initialize']

    def get_common_options_for_section(self, section: str) -> List[Flag]:
        if section not in self.configuration:
            return []

        options = []
        configuration_section = self.configuration[section]
        for flag in CONFIGURATION_FLAGS_DEFINITION[constants.SECTION_COMMON]:
            if flag in configuration_section:
                option = self.validate_configuration_option(
                    CONFIGURATION_FLAGS_DEFINITION[constants.SECTION_COMMON],
                    flag,
                    configuration_section[flag]
                )
                if option:
                    # special case for inherit flag
                    if option.key == constants.PARAMETER_INHERIT and option.value:
                        # run the common configuration for the parent
                        parent_options = self.get_common_options_for_section(option.value)
                        if parent_options:
                            options.extend(parent_options)

                    options.append(option)

        return options

    def validate_global_configuration_option(self, key: str, value: Union[str, int, bool, list]) -> Flag:
        '''
        Returns a validated global configuration flag (usually found in [global] section)
        '''
        return self.validate_configuration_option(GLOBAL_FLAGS_DEFINITION, key, value)

    def validate_configuration_option(self, definition: dict, key: str, value: Union[str, int, bool, list]) -> Flag:
        '''
        Validates against 'definition' and returns a configuration flag
        '''
        if key not in definition:
            return False

        if constants.DEFINITION_TYPE not in definition[key]:
            return False

        if isinstance(definition[key][constants.DEFINITION_TYPE], list):
            # this flag can be different types (exemple: boolean or string)
            for expected_type in definition[key][constants.DEFINITION_TYPE]:
                success = self.__check_type(expected_type, value, expect_list=(
                    'list' in definition[key] and definition[key]['list']))
                if success:
                    return self.__valid_flag(definition[key], key, value, expected_type)

        else:
            expected_type = definition[key][constants.DEFINITION_TYPE]
            success = self.__check_type(expected_type, value, expect_list=(
                'list' in definition[key] and definition[key]['list']))
            if success:
                return self.__valid_flag(definition[key], key, value, expected_type)

        return False

    def __check_type(self, expected_type: str, value: Union[str, int, bool, list], expect_list=False) -> bool:
        if expect_list and isinstance(value, list):
            for subvalue in value:
                success = self.__check_type(
                    expected_type, subvalue, expect_list=False)
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

    def __valid_flag(self, definition, key: str, value: Union[str, int, bool, list], expected_type: str) -> Flag:
        if constants.DEFINITION_FLAG in definition:
            # the restic flag has a different name than the configuration file flag
            key = definition[constants.DEFINITION_FLAG]
        return Flag(key, value, expected_type)
