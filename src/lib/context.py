from .config import defaults

class Context:

    def __init__(self):
        self.configuration_file = defaults['configuration_file']
        self.profile_name = defaults['profile_name']
        self.verbose = defaults['verbose']
        self.quiet = defaults['quiet']
