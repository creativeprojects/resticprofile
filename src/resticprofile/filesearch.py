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

    def find_file(self, filename: str, resolve=False) -> str:
        '''
        Returns a full file path from the configuration file location
        '''
        filepath = Path(filename)
        if filepath.is_absolute():
            return self._get_filepath(filepath, resolve)

        filepath = Path(self.configuration_directory) / filename
        return self._get_filepath(filepath, resolve)

    def find_dir(self, filename: str, resolve=False) -> str:
        '''
        Returns a directory path from the current active directory
        '''
        filepath = Path(filename)
        if filepath.is_absolute():
            return self._get_filepath(filepath, resolve)

        filepath = Path(getcwd()) / filename
        return self._get_filepath(filepath, resolve)

    def _get_filepath(self, filepath: Path, resolve=False) -> str:
        if resolve:
            return str(filepath.resolve())
        return str(filepath)
