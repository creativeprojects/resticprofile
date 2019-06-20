from typing import List

from resticprofile import constants

class Restic:
    def __init__(self, command=''):
        # set instance variables
        self.command = command
        self.repository = ""
        self.__common_arguments = [] # type: List[str]
        self.__arguments = {} # type: Dict[str, List[str]]
        self.backup_paths = []

    def get_init_command(self) -> str:
        return self.__get_command(constants.COMMAND_INIT)

    def get_check_command(self) -> str:
        return self.__get_command(constants.COMMAND_CHECK)

    def get_forget_command(self) -> str:
        return self.__get_command(constants.COMMAND_FORGET)

    def get_command(self) -> str:
        return self.__get_command(self.command)

    def __get_command(self, command_name: str) -> str:
        command = [command_name]
        if self.__common_arguments:
            command.extend(self.__common_arguments)

        if command_name in self.__arguments:
            command.extend(self.__arguments[command_name])

        if command_name == "backup" and self.backup_paths:
            command.extend(self.backup_paths)

        return ' '.join(command)

    def set_common_argument(self, arg):
        self.__common_arguments.append(arg)

    def extend_arguments(self, args: List[str]):
        if self.command not in self.__arguments:
            self.__arguments[self.command] = []
        self.__arguments[self.command].extend(args)
