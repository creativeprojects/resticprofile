from .config import configuration_flags, validate_configuration_option
from . import constants
from .flag import Flag

class Profile:

    def __init__(self, profile_name):
        self._new_profile(profile_name)

    def _new_profile(self, profile_name):
        self.quiet = None
        self.verbose = None
        self.profile_name = profile_name
        self.repository = ""
        self._global_flags = {} # type: Dict[str, Flag]
        self.source = []

    def set_global_configuration(self, configuration_section):
        if not configuration_section:
            return

        for flag in configuration_flags[constants.SECTION_GLOBAL]:
            if flag in configuration_section:
                option = validate_configuration_option(configuration_flags[constants.SECTION_GLOBAL], flag, configuration_section[flag])
                if option:
                    self.set_flag(option)

    def get_global_flags(self):
        flags = []
        # add the specific flags
        if self.repository:
            flags.extend(Flag('repo', self.repository, 'str').get_flags())
        flags.extend(Flag('quiet', self.quiet, 'bool').get_flags())
        if isinstance(self.verbose, bool):
            flags.extend(Flag('verbose', self.verbose, 'bool').get_flags())
        elif isinstance(self.verbose, int):
            flags.extend(Flag('verbose', self.verbose, 'int').get_flags())

        for _, flag in self._global_flags.items():
            # create a restic argument for it
            arguments = flag.get_flags()
            if arguments:
                flags.extend(arguments)
        return flags

    def set_flag(self, option):
        if not option: return
        if option.key == 'repo':
            if isinstance(option.value, str) and option.value:
                self.repository = option.value

        elif option.key == 'quiet':
            if isinstance(option.value, bool):
                self.quiet = option.value

        elif option.key == 'verbose':
            if isinstance(option.value, bool) or isinstance(option.value, int):
                self.verbose = option.value

        else:
            self._global_flags[option.key] = option

