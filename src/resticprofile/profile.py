from typing import List

from resticprofile import constants
from resticprofile.config import Config
from resticprofile.flag import Flag


class Profile:

    def __init__(self, config: Config, profile_name: str):
        self.quiet = None
        self.verbose = None
        self.config = config
        self.profile_name = profile_name
        self.inherit = None
        self.repository = ""
        self.initialize = False
        self.forget_before = False
        self.forget_after = False
        self.__common_flags = {}  # type: Dict[str, Flag]
        self.__command_flags = {}  # type: Dict[str, Dict[str, Flag]]
        self.__retention_flags = {}  # type: Dict[str, Flag]
        self.source = []

    def set_common_configuration(self):
        options = self.config.get_options_for_section(self.profile_name)

        if options:
            for option in options:
                self.__set_common_flag(option)

    def set_command_configuration(self, command: str):
        options = self.config.get_options_for_section(self.profile_name, command)

        if options:
            for option in options:
                self.__set_command_flag(option, command)

    def set_retention_configuration(self):
        options = self.config.get_options_for_retention(self.profile_name)

        if options:
            for option in options:
                self.__set_retention_flag(option)

    def get_global_flags(self) -> List[str]:
        flags = self.__get_specific_flags()
        for _, flag in self.__common_flags.items():
            # create a restic argument for it
            arguments = flag.get_flags()
            if arguments:
                flags.extend(arguments)

        return flags

    def get_command_flags(self, command: str) -> List[str]:
        if command not in self.__command_flags:
            return []

        flags = self.__get_specific_flags()
        for _, flag in self.__command_flags[command].items():
            # create a restic argument for it
            arguments = flag.get_flags()
            if arguments:
                flags.extend(arguments)

        if command == constants.COMMAND_BACKUP:
            flags.extend(self.get_backup_source())

        return flags

    def get_retention_flags(self) -> List[str]:
        path_not_present = True
        flags = self.get_global_flags()
        for key, flag in self.__retention_flags.items():
            # create a restic argument for it
            arguments = flag.get_flags()
            if arguments:
                flags.extend(arguments)
            # flag if the 'path' flag was specified
            if key == constants.PARAMETER_PATH:
                path_not_present = False

        # to make sure we only deal with the current backup, we add the backup source as 'path' argument
        if path_not_present and self.source:
            path_flag = Flag(constants.PARAMETER_PATH, self.source, 'dir')
            flags.extend(path_flag.get_flags())

        return flags

    def get_backup_source(self) -> List[str]:
        '''
        Returns a list of unique backup location
        '''
        sources = []
        for source in self.source:
            sources.append("'{}'".format(source.replace("'", "\'")))
        return list(set(sources))


    def __get_specific_flags(self) -> List[str]:
        flags = []
        # add the specific flags
        flags.extend(self.__get_repository_flag())
        flags.extend(self.__get_quiet_flag())
        flags.extend(self.__get_verbose_flag())
        return flags

    def __get_repository_flag(self) -> List[str]:
        if self.repository:
            return Flag(constants.PARAMETER_REPO, self.repository, 'str').get_flags()
        return []

    def __get_quiet_flag(self) -> List[str]:
        return Flag(constants.PARAMETER_QUIET, self.quiet, 'bool').get_flags()

    def __get_verbose_flag(self) -> List[str]:
        if isinstance(self.verbose, bool):
            return Flag(constants.PARAMETER_VERBOSE, self.verbose, 'bool').get_flags()
        elif isinstance(self.verbose, int):
            return Flag(constants.PARAMETER_VERBOSE, self.verbose, 'int').get_flags()
        return []

    def __set_common_flag(self, option: Flag):
        if not option:
            return
        if option.key == constants.PARAMETER_INHERIT:
            if isinstance(option.value, str) and option.value:
                self.inherit = option.value

        elif option.key == constants.PARAMETER_REPO:
            if isinstance(option.value, str) and option.value:
                self.repository = option.value

        elif option.key == constants.PARAMETER_QUIET:
            if isinstance(option.value, bool):
                self.quiet = option.value

        elif option.key == constants.PARAMETER_INITIALIZE:
            if isinstance(option.value, bool):
                self.initialize = option.value

        elif option.key == constants.PARAMETER_VERBOSE:
            if isinstance(option.value, bool) or isinstance(option.value, int):
                self.verbose = option.value

        # adds it to the list of flags
        if not self.__is_special_flag(option.key):
            self.__common_flags[option.key] = option

    def __set_command_flag(self, option: Flag, command: str):
        if not option:
            return

        # command specific flags
        if option.key == constants.PARAMETER_SOURCE:
            if isinstance(option.value, str) and option.value:
                self.source.append(option.value)
            elif isinstance(option.value, list) and option.value:
                self.source.extend(option.value)

        elif option.key == constants.PARAMETER_INITIALIZE:
            if isinstance(option.value, bool):
                self.initialize = option.value

        if command not in self.__command_flags:
            self.__command_flags[command] = {}

        if not self.__is_special_flag(option.key):
            self.__command_flags[command][option.key] = option


    def __set_retention_flag(self, option: Flag):
        if not option:
            return
        if option.key == constants.PARAMETER_FORGET_BEFORE_BACKUP:
            if isinstance(option.value, bool):
                self.forget_before = option.value

        elif option.key == constants.PARAMETER_FORGET_AFTER_BACKUP:
            if isinstance(option.value, bool):
                self.forget_after = option.value

        # adds it to the list of flags
        if not self.__is_special_flag(option.key):
            self.__retention_flags[option.key] = option


    def __is_special_flag(self, key: str):
        return key in (
            constants.PARAMETER_REPO,
            constants.PARAMETER_QUIET,
            constants.PARAMETER_VERBOSE,
            constants.PARAMETER_INHERIT,
            constants.PARAMETER_SOURCE,
            constants.PARAMETER_INITIALIZE,
            constants.PARAMETER_FORGET_BEFORE_BACKUP,
            constants.PARAMETER_FORGET_AFTER_BACKUP,
        )
