from .config import restic_flags, validate_configuration_option
from . import constants

class Backup:

    def __init__(self, profile_name):
        self.quiet = None
        self.verbose = None
        self.profile_name = profile_name
        self.flags = []
        self.source = []

    def set_global_configuration(self, configuration_section):

        if not configuration_section:
            return

        for flag in restic_flags[constants.SECTION_GLOBAL]:
            if flag in configuration_section:
                result = validate_configuration_option(restic_flags[constants.SECTION_GLOBAL], flag, configuration_section[flag])
                argument = self.get_argument(result)
                if argument:
                    self.flags.append(argument)

    def get_argument(self, result):
        if not result: return ''
        if result['type'] == 'bool':
            if result['value']:
                return "--{}".format(result['key'])
            else:
                return ''
        else:
            return "--{}={}".format(result['key'], result['value'])
