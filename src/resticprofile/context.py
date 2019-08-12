'''
resticprofile global context
'''
from getopt import getopt, GetoptError

from resticprofile import constants
from resticprofile.config import Config
from resticprofile.filesearch import find_restic_binary, get_restic_binary
from resticprofile.console import Console

class Context:

    def __init__(self, arguments_definition: dict):
        self.arguments_definition = arguments_definition
        self.restic_path = None
        self.configuration_file = constants.DEFAULT_CONFIGURATION_FILE
        self.profile_name = constants.DEFAULT_PROFILE_NAME
        self.default_command = constants.DEFAULT_COMMAND
        self.initialize = constants.DEFAULT_INITIALIZE_FLAG
        self.verbose = constants.DEFAULT_VERBOSE_FLAG
        self.quiet = constants.DEFAULT_QUIET_FLAG
        self.ansi = True
        self.nice = None
        self.ionice = None
        self.opts = []
        self.args = []

    def load_context_from_command_line(self, argv: list):
        try:
            short_options = self._get_short_options()
            long_options = self._get_long_options()
            self.opts, self.args = getopt(argv[1:], short_options, long_options)

        except GetoptError as err:
            console = Console()
            console.error("Error in the command arguments: " + err.msg)
            console.usage(argv[0])
            exit(2)

        for option, argument in self.opts:
            if option in self._get_possible_options_for('help'):
                Console().usage(argv[0])
                exit()

            elif option in self._get_possible_options_for('quiet'):
                self.quiet = True

            elif option in self._get_possible_options_for('verbose'):
                self.verbose = True

            elif option in self._get_possible_options_for('config'):
                self.configuration_file = argument

            elif option in self._get_possible_options_for('name'):
                self.profile_name = argument

            elif option in self._get_possible_options_for('no-ansi'):
                self.ansi = False

            else:
                assert False, "unhandled option"

    def _get_short_options(self):
        short_options = ""
        for _, options in self.arguments_definition.items():
            if 'short' in options:
                short_options += options['short'] + (":" if options['argument'] else "")
        return short_options


    def _get_long_options(self):
        long_options = []
        for _, options in self.arguments_definition.items():
            if 'long' in options:
                long_options.append(options['long'] + ("=" if options['argument'] else ""))
        return long_options


    def _get_possible_options_for(self, option):
        options = []
        if 'short' in self.arguments_definition[option]:
            options.append("-{}".format(self.arguments_definition[option]['short']))
        if 'long' in self.arguments_definition[option]:
            options.append("--{}".format(self.arguments_definition[option]['long']))
        return options


    def set_global_context(self, config: Config):
        self.ionice = config.get_ionice()
        self.nice = config.get_nice()
        self.default_command = config.get_default_command()
        self.initialize = config.get_initialize()
        self.restic_path = self._quote_path(config.get_restic_binary_path())

    def get_restic_path(self):
        if self.restic_path:
            return self.restic_path

        self.restic_path = find_restic_binary()
        if not self.restic_path:
            # if all fails, the shell might be able to find it?
            self.restic_path = get_restic_binary()

        self.restic_path = self._quote_path(self.restic_path)

        return self.restic_path

    def get_command_prefix(self) -> str:
        command_prefix = ""
        if self.nice:
            command_prefix += self.nice.get_command() + ' '
        if self.ionice:
            command_prefix += self.ionice.get_command() + ' '

        return command_prefix

    def _quote_path(self, path: str) -> str:
        if path and path.find(' ') > -1:
            path = '"{}"'.format(path)
        return path
