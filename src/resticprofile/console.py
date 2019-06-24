'''
Display messages to the console
'''
from colorama import Fore, Back, Style, init
from resticprofile import constants
from resticprofile.help import get_options_help


class Console:
    def __init__(self, quiet=False, verbose=False):
        '''
        Display messages to the console
        '''
        self.quiet = quiet
        self.verbose = verbose
        init(autoreset=True)

    def debug(self, message):
        '''
        Display debug message to the console
        '''
        if self.verbose:
            print(Fore.LIGHTGREEN_EX + message)

    def info(self, message):
        '''
        Display info message to the console
        '''
        if not self.quiet:
            print(Fore.LIGHTYELLOW_EX + message)

    def warning(self, message):
        '''
        Display warning message to the console
        '''
        print(Fore.LIGHTRED_EX + message)

    def error(self, message):
        '''
        Display error message to the console
        '''
        print(Fore.LIGHTRED_EX + message)

    def usage(self, name):
        '''
        Display usage to the console
        '''
        print("\nUsage:")
        print(" " + name + "\n   " + "\n   ".join(get_options_help(constants.ARGUMENTS_DEFINITION)) + "\n   [command]\n")
        print("Default configuration file is: '{}' (in the current folder)".format(constants.DEFAULT_CONFIGURATION_FILE))
        print("Default configuration profile is: {}\n".format(constants.DEFAULT_PROFILE_NAME))
