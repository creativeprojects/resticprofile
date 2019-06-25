'''
Class for ionice command
'''
import platform


class IONice:
    def __init__(self, io_class=2, io_level=4, ignore_failure=False):
        if io_class not in (0, 1, 2, 3):
            io_class = 2
        self.io_class = io_class

        if io_level > 7:
            io_level = 7
        elif io_level < 0:
            io_level = 0
        self.io_level = io_level
        self.ignore_failure = ignore_failure

    def get_command(self, system=None):
        if not system:
            system = platform.system()

        if system == "Linux":
            command = "ionice -c {}".format(self.io_class)
            if self.io_class in (1, 2):
                command += " -n {}".format(self.io_level)
            if self.ignore_failure:
                command += " -t"

            return command

        if not self.ignore_failure:
            raise Exception("'ionice' is not available on {}. ".format(system) + 
                            "Please raise a defect on github if you think it's available for your platform.")

        return ""
