from os import getcwd
from pathlib import Path

DEFAULT_SEARCH_LOCATIONS = [
    '/usr/local/etc/',
    '/etc/',
]


def find_configuration_file(configuration_file: str) -> str:
    '''
    Search for the file in the current directory, the home directory, and the locations specified in DEFAULT_SEARCH_LOCATIONS
    Returns None if the file was not found
    '''
    for filepath in list(
            map(
                lambda path: Path(path) / configuration_file,
                [getcwd(), str(Path().home())] + DEFAULT_SEARCH_LOCATIONS
            )
        ):
        if filepath.is_file():
            return str(filepath)

    return None


class FileSearch:

    def __init__(self, configuration_directory: str):
        self.configuration_directory = configuration_directory

    def find_file(self, filename: str) -> str:
        '''
        Returns a full file path from the configuration file location
        '''
        if Path(filename).is_absolute():
            return filename

        filepath = Path(self.configuration_directory) / filename
        return str(filepath)

    def find_dir(self, filename: str) -> str:
        '''
        Returns a directory path from the current active directory
        '''
        if Path(filename).is_absolute():
            return filename

        filepath = Path(filename)
        return str(filepath)
