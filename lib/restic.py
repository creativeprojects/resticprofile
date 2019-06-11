
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
        return "init " + ' '.join(self.common_arguments)

    def get_prune_command(self):
        return "prune " + ' '.join(self.common_arguments)
    
    def get_command(self):
        return self.command + " " + ' '.join(self.common_arguments) + " " + ' '.join(self.arguments[self.command]) + " " + ' '.join(self.backup_paths)

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
