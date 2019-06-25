class ConfigError(Exception):
    """Exception raised for errors in the configuration file.

    Attributes:
        section -- section where the error occurred
        message -- explanation of the error
    """

    def __init__(self, section: str, message: str):
        self.section = section
        self.message = message
