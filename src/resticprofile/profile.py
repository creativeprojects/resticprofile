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
        self.__common_flags = {}  # type: Dict[str, Flag]
        self.source = []

    def set_common_configuration(self):
        options = self.config.get_common_options_for_section(self.profile_name)

        if options:
            for option in options:
                self.set_flag(option)

    def get_global_flags(self):
        flags = []
        # add the specific flags
        flags.extend(self.__get_repository_flag())
        flags.extend(self.__get_quiet_flag())
        flags.extend(self.__get_verbose_flag())

        for _, flag in self.__common_flags.items():
            # create a restic argument for it
            arguments = flag.get_flags()
            if arguments:
                flags.extend(arguments)
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

    def set_flag(self, option):
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

        elif option.key == constants.PARAMETER_VERBOSE:
            if isinstance(option.value, bool) or isinstance(option.value, int):
                self.verbose = option.value

        # adds it to the list of flags
        if option.key not in (
                constants.PARAMETER_REPO,
                constants.PARAMETER_QUIET,
                constants.PARAMETER_VERBOSE,
                constants.PARAMETER_INHERIT
            ):
            self.__common_flags[option.key] = option
