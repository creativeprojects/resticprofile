'''
Generate restic command line
'''
from typing import List

from resticprofile import constants

class Restic:
    def __init__(self, command=''):
        # set instance variables
        self.command = command
        self.repository = ""
        self._common_arguments = [] # type: List[str]
        self._arguments = {} # type: Dict[str, List[str]]
        self.backup_paths = []

    def get_init_command(self) -> str:
        return self._get_command(constants.COMMAND_INIT)

    def get_check_command(self) -> str:
        return self._get_command(constants.COMMAND_CHECK)

    def get_forget_command(self) -> str:
        return self._get_command(constants.COMMAND_FORGET)

    def get_command(self) -> str:
        return self._get_command(self.command)

    def _get_command(self, command_name: str) -> str:
        command = [command_name]
        if self._common_arguments:
            command.extend(self._common_arguments)

        if command_name in self._arguments:
            command.extend(self._arguments[command_name])

        if command_name == "backup" and self.backup_paths:
            command.extend(self.backup_paths)

        return ' '.join(command)

    def set_common_argument(self, arg):
        self._common_arguments.append(arg)

    def extend_arguments(self, args: List[str]):
        if self.command not in self._arguments:
            self._arguments[self.command] = []
        self._arguments[self.command].extend(args)
