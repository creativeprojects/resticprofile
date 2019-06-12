from colorama import Fore, Back, Style, init

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
