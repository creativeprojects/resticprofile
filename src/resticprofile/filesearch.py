from os import getcwd
from pathlib import Path

DEFAULT_SEARCH_LOCATIONS = [
    '~/',
    '/usr/local/etc/',
    '/etc/',
]


def find_configuration_file(configuration_file: str) -> str:
    for filepath in list(
            map(
                lambda path: Path(path + configuration_file),
                [getcwd() + "/"] + DEFAULT_SEARCH_LOCATIONS
            )
        ):
        if filepath.is_file():
            return str(filepath)

    return None


class FileSearch:

    def __init__(self, configuration_directory: str):
        self.configuration_directory = configuration_directory

    def find(self, filename: str) -> str:
        '''
        Returns a full file path of the filename
        '''
        if Path(filename).is_absolute():
            return filename

        filepath = Path(self.configuration_directory) / filename
        return str(filepath)
