from .config import restic_flags, validate_configuration_option
from . import constants

class Profile:

    def __init__(self, profile_name):
        self._new_profile(profile_name)

    def _new_profile(self, profile_name):
        self.quiet = None
        self.verbose = None
        self.profile_name = profile_name
        self.repository = ""
        self.global_flags = []
        self.command_flags = []
        self.source = []

    def set_global_configuration(self, configuration_section):
        if not configuration_section:
            return

        for flag in restic_flags[constants.SECTION_GLOBAL]:
            if flag in configuration_section:
                result = validate_configuration_option(restic_flags[constants.SECTION_GLOBAL], flag, configuration_section[flag])
                # populate the context with special flags (like reposiroty and such)
                self.set_context_of_special_flag(result)
                # then create a restic argument for it
                arguments = self.get_flags(result)
                if arguments:
                    self.global_flags.extend(arguments)

    def get_flags(self, result):
        if not result: return ''

        flags = []
        if isinstance(result['value'], list):
            for value in result['value']:
                flag = self.get_single_flag(result['key'], value, result['type'])
                if flag:
                    flags.append(flag)

        else:
            # still return a list but with one element
            flag = self.get_single_flag(result['key'], result['value'], result['type'])
            if flag:
                flags.append(flag)

        return flags

    def get_single_flag(self, key, value, type_value):
        if type_value == 'bool':
            if value:
                return "--{}".format(key)
            else:
                return ''
        else:
            return "--{} {}".format(key, value)

    def set_context_of_special_flag(self, option):
        if not option: return
        if option['key'] == 'repo':
            if isinstance(option['value'], str) and option['value']:
                self.repository = option['value']

        elif option['key'] == 'quiet':
            if isinstance(option['value'], bool):
                self.quiet = option['value']

        elif option['key'] == 'verbose':
            if isinstance(option['value'], bool) or isinstance(option['value'], int):
                self.verbose = option['value']

