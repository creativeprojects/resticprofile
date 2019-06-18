from .config import Config
from . import constants
from .flag import Flag

class Profile:

    def __init__(self, config: Config, profile_name: str):
        self.quiet = None
        self.verbose = None
        self.config = config
        self.profile_name = profile_name
        self.repository = ""
        self.__common_flags = {} # type: Dict[str, Flag]
        self.source = []

    def set_common_configuration(self):
        options = self.config.get_common_options_for_section(self.profile_name)

        if options:
            for option in options:
                self.set_flag(option)

    def get_global_flags(self):
        flags = []
        # add the specific flags
        if self.repository:
            flags.extend(Flag(constants.PARAMETER_REPO, self.repository, 'str').get_flags())
        flags.extend(Flag(constants.PARAMETER_QUIET, self.quiet, 'bool').get_flags())
        if isinstance(self.verbose, bool):
            flags.extend(Flag(constants.PARAMETER_VERBOSE, self.verbose, 'bool').get_flags())
        elif isinstance(self.verbose, int):
            flags.extend(Flag(constants.PARAMETER_VERBOSE, self.verbose, 'int').get_flags())

        for _, flag in self.__common_flags.items():
            # create a restic argument for it
            arguments = flag.get_flags()
            if arguments:
                flags.extend(arguments)
        return flags

    def set_flag(self, option):
        if not option:
            return
        if option.key == constants.PARAMETER_REPO:
            if isinstance(option.value, str) and option.value:
                self.repository = option.value

        elif option.key == constants.PARAMETER_QUIET:
            if isinstance(option.value, bool):
                self.quiet = option.value

        elif option.key == constants.PARAMETER_VERBOSE:
            if isinstance(option.value, bool) or isinstance(option.value, int):
                self.verbose = option.value

        # adds it to the list of flags
        if option.key not in (constants.PARAMETER_REPO, constants.PARAMETER_QUIET, constants.PARAMETER_VERBOSE):
            self.__common_flags[option.key] = option
