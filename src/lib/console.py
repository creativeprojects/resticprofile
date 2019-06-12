from colorama import Fore, Back, Style, init
from .config import defaults, arguments_definition
from .help import get_options_help

class Console:
    def __init__(self, quiet = False, verbose = False):
        self.quiet = quiet
        self.verbose = verbose
        init(autoreset = True)

    def debug(self, message):
        if (self.verbose):
            print(Fore.LIGHTGREEN_EX + message)

    def info(self, message):
        if (not self.quiet):
            print(Fore.LIGHTYELLOW_EX + message)
    
    def warning(self, message):
        print(Fore.LIGHTRED_EX + message)

    def error(self, message):
        print(Fore.LIGHTRED_EX + message)

    def usage(self, name):
        print("\nUsage:")
        print(" " + name + "\n   " + "\n   ".join(get_options_help(arguments_definition)) + "\n   [command]\n")
        print
        print("Default configuration file is: '{}' (in the current folder)".format(defaults['configuration_file']))
        print("Default configuration profile is: {}\n".format(defaults['profile_name']))
