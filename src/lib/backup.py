from .config import restic_flags, validate_configuration_option
from . import constants
from .profile import Profile

class Backup(Profile):

    def __init__(self, profile_name):
        self._new_profile(profile_name)
        self.source = []
