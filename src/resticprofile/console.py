'''
Display messages to the console
'''
import time
from enum import Enum
from colorama import Fore, init
from resticprofile import constants
from resticprofile.help import get_options_help


class ConsoleColouring(Enum):
    NONE = 0
    DEFAULT = 1
    BRIGHT = 2


ANSI_OUTPUT = {
    ConsoleColouring.DEFAULT: {
        'debug': Fore.GREEN,
        'info': Fore.YELLOW,
        'warning': Fore.RED,
        'error': Fore.RED,
    },
    ConsoleColouring.BRIGHT: {
        'debug': Fore.LIGHTGREEN_EX,
        'info': Fore.LIGHTYELLOW_EX,
        'warning': Fore.LIGHTRED_EX,
        'error': Fore.LIGHTRED_EX,
    },
}

class Console:
    def __init__(self, quiet=False, verbose=False, ansi=True):
        '''
        Display messages to the console
        '''
        self.quiet = quiet
        self.verbose = verbose
        if ansi:
            self.colouring = ConsoleColouring.DEFAULT
        else:
            self.colouring = ConsoleColouring.NONE
        init(autoreset=True)

    def _msg(self, message, ansi=None):
        '''
        Low level display message
        '''
        timed_message = time.asctime() + ' ' + message
        if ansi:
            print(ansi + timed_message)
        else:
            print(timed_message)

    def _get_ansi(self, message_type: str):
        if self.colouring.value and self.colouring in ANSI_OUTPUT:
            return ANSI_OUTPUT[self.colouring][message_type]
        return ''

    def debug(self, message):
        '''
        Display debug message to the console
        '''
        if self.verbose:
            self._msg(message, self._get_ansi('debug'))

    def info(self, message):
        '''
        Display info message to the console
        '''
        if not self.quiet:
            self._msg(message, self._get_ansi('info'))

    def warning(self, message):
        '''
        Display warning message to the console
        '''
        self._msg(message, self._get_ansi('warning'))

    def error(self, message):
        '''
        Display error message to the console
        '''
        self._msg(message, self._get_ansi('error'))

    def usage(self, name):
        '''
        Display usage to the console
        '''
        print("\nUsage:")
        print(" " + name + "\n   " + "\n   ".join(get_options_help(constants.ARGUMENTS_DEFINITION)))
        print("   [restic command] [additional parameters to pass to restic]\n")
