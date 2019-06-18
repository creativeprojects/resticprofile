from .config import DEFAULTS, Config
from .flag import Flag
from .ionice import IONice
from .nice import Nice
from . import constants

class Context:

    def __init__(self):
        self.configuration_file = DEFAULTS['configuration_file']
        self.profile_name = DEFAULTS['profile_name']
        self.default_command = DEFAULTS['default_command']
        self.initialize = DEFAULTS['initialize']
        self.verbose = DEFAULTS['verbose']
        self.quiet = DEFAULTS['quiet']
        self.nice = None
        self.ionice = None

    def set_global_context(self, config: Config):
        self.ionice = config.get_ionice()
        self.nice = config.get_nice()
        self.default_command = config.get_default_command()
        self.initialize = config.get_initialize()
