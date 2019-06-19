
class Restic:
    def __init__(self):
        # set instance variables
        self.command = ""
        self.repository = ""
        self.common_arguments = []
        self.arguments = {}
        self.backup_paths = []
        self.prune_before = False
        self.prune_after = False

    def get_init_command(self):
        return self._get_command("init")

    def get_prune_command(self):
        return self._get_command("prune")

    def get_command(self):
        return self._get_command(self.command)

    def _get_command(self, command_name):
        command = [command_name]
        if self.common_arguments:
            command.extend(self.common_arguments)

        if command_name in self.arguments:
            command.extend(self.arguments[command_name])

        if command_name == "backup" and self.backup_paths:
            command.extend(self.backup_paths)

        return ' '.join(command)

    def set_common_argument(self, arg):
        self.common_arguments.append(arg)

    def set_argument(self, arg):
        if self.command not in self.arguments:
            self.arguments[self.command] = []
        self.arguments[self.command].append(arg)

    def extend_arguments(self, args):
        if self.command not in self.arguments:
            self.arguments[self.command] = []
        self.arguments[self.command].extend(args)
