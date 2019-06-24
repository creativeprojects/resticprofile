'''
Backup profile groups
'''
from resticprofile import constants

class Groups:
    def __init__(self, configuration: dict):
        self._groups = {} # type: Dist[string, list]
        self._configuration = configuration
        self._load_groups()

    def _load_groups(self):
        if constants.SECTION_CONFIGURATION_GROUPS in self._configuration:
            for group_name, profiles in self._configuration[constants.SECTION_CONFIGURATION_GROUPS].items():
                if group_name and profiles and isinstance(profiles, list):
                    self._groups[group_name] = profiles

    def exists(self, group_name: str):
        return group_name in self._groups

    def get_profiles(self, group_name: str):
        return self._groups[group_name]
