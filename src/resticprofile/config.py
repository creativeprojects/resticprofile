'''
resticprofile configuration
'''
from typing import Union, List, Dict
from resticprofile import constants
from resticprofile.flag import Flag
from resticprofile.ionice import IONice
from resticprofile.nice import Nice
from resticprofile.filesearch import FileSearch
from resticprofile.error import ConfigError


GLOBAL_FLAGS_DEFINITION = {
    'ionice': {'type': 'bool'},
    'ionice-class': {'type': 'int'},
    'ionice-level': {'type': 'int'},
    'nice': {'type': ['bool', 'int']},
    'default-command': {'type': 'str'},
    'initialize': {'type': 'bool'},
    'restic-binary': {'type': 'str'},
}

CONFIGURATION_FLAGS_DEFINITION = {
    'common': {
        'inherit': {'type': 'str'},
        'initialize': {'type': 'bool'},
        'cacert': {'type': 'file'},
        'cache-dir': {'type': 'file'},  # keep it as a 'file' type
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
    'retention': {
        'before-backup': {'type': 'bool'},
        'after-backup': {'type': 'bool'},
    },
    'backup': {
        'run-before': {'type': 'str', 'list': True},
        'run-after': {'type': 'str', 'list': True},
        'check-before': {'type': 'bool'},
        'check-after': {'type': 'bool'},
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
        'stdin-filename': {'type': 'str'},
        'tag': {'type': 'str', 'list': True},
        'time': {'type': 'str'},
        'with-atime': {'type': 'bool'},
        'source': {'type': 'dir', 'list': True},
    },
    'snapshots': {
        'compact': {'type': 'bool'},
        'group-by': {'type': 'str'},
        'host': {'type': ['bool', 'str']},
        'last': {'type': 'bool'},
        'path': {'type': 'dir', 'list': True},
        'tag': {'type': 'str', 'list': True},
    },
    "forget": {
        'keep-last': {'type': 'int'},
        'keep-hourly': {'type': 'int'},
        'keep-daily': {'type': 'int'},
        'keep-weekly': {'type': 'int'},
        'keep-monthly': {'type': 'int'},
        'keep-yearly': {'type': 'int'},
        'keep-within': {'type': 'str'},
        'keep-tag': {'type': 'str', 'list': True},
        'host': {'type': ['bool', 'str'], 'default': True},
        'tag': {'type': 'str', 'list': True},
        'path': {'type': 'dir', 'list': True},
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
    "mount": {
        'allow-other': {'type': 'bool'},
        'allow-root': {'type': 'bool'},
        'host': {'type': ['bool', 'str']},
        'no-default-permissions': {'type': 'bool'},
        'owner-root': {'type': 'bool'},
        'path': {'type': 'str', 'list': True},
        'snapshot-template': {'type': 'str'},
        'tag': {'type': 'str', 'list': True},
    }
}


class Config:
    '''
    Manage configuration information from configuration dictionnary
    '''

    def __init__(self, configuration: dict, file_search: FileSearch):
        self.configuration = configuration
        self.file_search = file_search

    def get_ionice(self, section=constants.SECTION_CONFIGURATION_GLOBAL) -> IONice:
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

    def get_nice(self, section=constants.SECTION_CONFIGURATION_GLOBAL) -> Nice:
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

    def get_default_command(self, section=constants.SECTION_CONFIGURATION_GLOBAL) -> str:
        if section not in self.configuration:
            return constants.DEFAULT_COMMAND

        configuration = self.configuration[section]
        if constants.PARAMETER_DEFAULT_COMMAND in configuration:
            default_command = self.validate_global_configuration_option(
                constants.PARAMETER_DEFAULT_COMMAND,
                configuration[constants.PARAMETER_DEFAULT_COMMAND]
            )
            if default_command:
                return default_command.value

        return constants.DEFAULT_COMMAND

    def get_initialize(self, section=constants.SECTION_CONFIGURATION_GLOBAL) -> bool:
        if section not in self.configuration:
            return constants.DEFAULT_INITIALIZE_FLAG

        configuration = self.configuration[section]
        if constants.PARAMETER_INITIALIZE in configuration:
            initialize = self.validate_global_configuration_option(
                constants.PARAMETER_INITIALIZE,
                configuration[constants.PARAMETER_INITIALIZE]
            )
            if initialize:
                return initialize.value

        return constants.DEFAULT_INITIALIZE_FLAG

    def get_restic_binary_path(self, section=constants.SECTION_CONFIGURATION_GLOBAL) -> bool:
        if section not in self.configuration:
            return None

        configuration = self.configuration[section]
        if constants.PARAMETER_RESTIC_BINARY in configuration:
            path = self.validate_global_configuration_option(
                constants.PARAMETER_RESTIC_BINARY,
                configuration[constants.PARAMETER_RESTIC_BINARY]
            )
            if path:
                return path.value

        return None

    def get_options_for_section(self, section: str, command='') -> List[Flag]:
        '''
        Returns the list of flags for the section.
        With no command parameter, it returns the common section of the profile
        With a command parameter, it returns the command section of the profile
        '''
        # common section
        options, inherit = self._get_options_for_common_section(section, [])
        # command section
        if command:
            inherit.reverse()
            for inherit_profile in inherit:
                inherited_common_options, _ = self._get_options_for_common_section(inherit_profile, [])
                options.extend(inherited_common_options)
                options.extend(self._get_options_for_command_section(inherit_profile, command))

            options.extend(self._get_options_for_command_section(section, command))

        return options

    def _get_options_for_common_section(self, section: str, inherit: List[str]) -> (List[Flag], List[str]):
        if section not in self.configuration:
            return ([], inherit)

        section_definition = constants.SECTION_DEFINITION_COMMON
        options = []
        configuration_section = self.configuration[section]
        for flag in CONFIGURATION_FLAGS_DEFINITION[section_definition]:
            if flag in configuration_section:
                option = self.validate_configuration_option(
                    CONFIGURATION_FLAGS_DEFINITION[section_definition],
                    flag,
                    configuration_section[flag]
                )
                if option:
                    # special case for inherit flag
                    if option.key == constants.PARAMETER_INHERIT:
                        if option.value:
                            inherit.append(option.value)
                            # run the common configuration for the parent
                            parent_options, inherit = self._get_options_for_common_section(option.value, inherit)
                            if parent_options:
                                options.extend(parent_options)
                    else:
                        options.append(option)

        return (options, inherit)

    def _get_options_for_command_section(self, section: str, command: str) -> List[Flag]:
        if section not in self.configuration:
            return []

        section_definition = command
        if section_definition not in CONFIGURATION_FLAGS_DEFINITION:
            # unknown command
            return []

        options = []
        configuration_section = self.configuration[section]
        if command and command in configuration_section:
            configuration_section = configuration_section[command]

        # configuration flags can also be common flags, so we merge both sections (the dict merge syntax is kinda weird)
        configuration_flags_definition = {**CONFIGURATION_FLAGS_DEFINITION[section_definition], \
            **CONFIGURATION_FLAGS_DEFINITION[constants.SECTION_DEFINITION_COMMON]}

        for flag in configuration_flags_definition:
            if flag in configuration_section:
                option = self.validate_configuration_option(
                    configuration_flags_definition,
                    flag,
                    configuration_section[flag]
                )
                if option:
                    options.append(option)

        return options

    def get_options_for_retention(self, section: str) -> List[Flag]:
        # common section
        options, inherit = self._get_options_for_common_section(section, [])
        # retention section
        inherit.reverse()
        for inherit_profile in inherit:
            inherited_common_options, _ = self._get_options_for_common_section(inherit_profile, [])
            options.extend(inherited_common_options)
            options.extend(self._get_options_for_retention(inherit_profile))

        options.extend(self._get_options_for_retention(section))

        return options

    def _get_options_for_retention(self, section: str) -> List[Flag]:
        if section not in self.configuration:
            return []

        configuration_section = self.configuration[section]
        if constants.SECTION_CONFIGURATION_RETENTION not in configuration_section:
            return []
        configuration_section = configuration_section[constants.SECTION_CONFIGURATION_RETENTION]

        # configuration flags are the specific ones to 'retention' + the ones from the 'forget' command
        configuration_flags_definition = {**CONFIGURATION_FLAGS_DEFINITION[constants.SECTION_CONFIGURATION_RETENTION], \
            **CONFIGURATION_FLAGS_DEFINITION[constants.SECTION_DEFINITION_FORGET]}

        options = []
        for flag in configuration_flags_definition:
            if flag in configuration_section:
                option = self.validate_configuration_option(
                    configuration_flags_definition,
                    flag,
                    configuration_section[flag]
                )
                if option:
                    options.append(option)

        return options

    def get_environment(self, section: str) -> Dict[str, str]:
        env = {}
        # common section
        _, inherit = self._get_options_for_common_section(section, [])
        inherit.reverse()
        for inherit_profile in inherit:
            if not inherit_profile in self.configuration:
                raise ConfigError(section, "Inherited profile [{}] was not found in the configuration".format(inherit_profile))

            if constants.SECTION_CONFIGURATION_ENVIRONMENT in self.configuration[inherit_profile]:
                env_config = self.configuration[inherit_profile][constants.SECTION_CONFIGURATION_ENVIRONMENT]
                for key in env_config:
                    env[key.upper()] = env_config[key]

        if constants.SECTION_CONFIGURATION_ENVIRONMENT in self.configuration[section]:
            env_config = self.configuration[section][constants.SECTION_CONFIGURATION_ENVIRONMENT]
            for key in env_config:
                env[key.upper()] = env_config[key]

        return env

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
                success = self._check_type(expected_type, value, expect_list=('list' in definition[key] and definition[key]['list']))
                if success:
                    return self._valid_flag(definition[key], key, value, expected_type)

        else:
            expected_type = definition[key][constants.DEFINITION_TYPE]
            success = self._check_type(expected_type, value, expect_list=('list' in definition[key] and definition[key]['list']))
            if success:
                return self._valid_flag(definition[key], key, value, expected_type)

        return False

    def _check_type(self, expected_type: str, value: Union[str, int, bool, list], expect_list=False) -> bool:
        if expect_list and isinstance(value, list):
            for subvalue in value:
                success = self._check_type(
                    expected_type, subvalue, expect_list=False)
                if not success:
                    return False
            return True

        if expected_type == 'bool':
            return isinstance(value, bool)
        elif expected_type == 'int':
            return isinstance(value, int)
        elif expected_type in ('str', 'file', 'dir'):
            return isinstance(value, str)
        else:
            raise Exception("Unknown type '{}'".format(expected_type))

    def _valid_flag(self, definition, key: str, value: Union[str, int, bool, list], expected_type: str) -> Flag:
        if constants.DEFINITION_FLAG in definition:
            # the restic flag has a different name than the configuration file flag
            key = definition[constants.DEFINITION_FLAG]

        if value:
            if expected_type == 'file':
                value = self._get_file_value(value)
            elif expected_type == 'dir':
                value = self._get_dir_value(value)

        return Flag(key, value, expected_type)

    def _get_file_value(self, value: Union[str, list]):
        if isinstance(value, str):
            parsed_value = self.file_search.find_file(value)
        elif isinstance(value, list):
            parsed_value = []
            for single_value in value:
                parsed_value.append(self.file_search.find_file(single_value))

        return parsed_value

    def _get_dir_value(self, value: Union[str, list]):
        if isinstance(value, str):
            parsed_value = self.file_search.find_dir(value)
        elif isinstance(value, list):
            parsed_value = []
            for single_value in value:
                # this method can throw a file not found
                parsed_value.append(self.file_search.find_dir(single_value))

        return parsed_value
