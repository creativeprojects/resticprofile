from .config import defaults, global_flags, validate_configuration_option
from .ionice import IONice
from .nice import Nice

class Context:

    def __init__(self):
        self.configuration_file = defaults['configuration_file']
        self.profile_name = defaults['profile_name']
        self.default_command = defaults['default_command']
        self.initialize = defaults['initialize']
        self.verbose = defaults['verbose']
        self.quiet = defaults['quiet']
        self.nice = None
        self.ionice = None

    def set_global_context(self, configuration_section):

        if not configuration_section:
            return

        self.set_ionice(configuration_section)
        self.set_nice(configuration_section)
        self.set_default_command(configuration_section)
        self.set_initialize(configuration_section)

    def set_ionice(self, configuration_section):
        if 'ionice' in configuration_section:
            ionice = validate_configuration_option(global_flags, 'ionice', configuration_section['ionice'])
            if ionice and ionice['value']:
                io_class = None
                io_level = None
                if 'ionice-class' in configuration_section:
                    io_class = validate_configuration_option(global_flags, 'ionice-class', configuration_section['ionice-class'])
                if 'ionice-level' in configuration_section:
                    io_level = validate_configuration_option(global_flags, 'ionice-level', configuration_section['ionice-level'])

                if io_class and io_level:
                    self.ionice = IONice(io_class['value'], io_level['value'])
                elif io_class:
                    self.ionice = IONice(io_class = io_class['value'])
                elif io_level:
                    self.ionice = IONice(io_level = io_level['value'])
                else:
                    self.ionice = IONice()

    def set_nice(self, configuration_section):
        if 'nice' in configuration_section:
            nice = validate_configuration_option(global_flags, 'nice', configuration_section['nice'])
            if nice and (nice['value'] and nice['type'] == 'bool') or (nice['type'] == 'int'):
                if nice['type'] == 'int':
                    self.nice = Nice(nice['value'])
                else:
                    self.nice = Nice()

    def set_default_command(self, configuration_section):
        if 'default-command' in configuration_section:
            default_command = validate_configuration_option(global_flags, 'default-command', configuration_section['default-command'])
            if default_command:
                self.default_command = default_command['value']

    def set_initialize(self, configuration_section):
        if 'initialize' in configuration_section:
            initialize = validate_configuration_option(global_flags, 'initialize', configuration_section['initialize'])
            if initialize:
                self.initialize = initialize['value']
