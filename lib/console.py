from colorama import Fore, Back, Style, init

class Console:
    def __init__(self, quiet = False, verbose = False):
        self.quiet = quiet
        self.verbose = verbose
        init(autoreset = True)

    def debug(self, message):
        if (self.verbose):
            print(Fore.GREEN + message)

    def info(self, message):
        if (not self.quiet):
            print(Fore.YELLOW + message)
    
    def error(self, message):
        print(Fore.RED + message)
