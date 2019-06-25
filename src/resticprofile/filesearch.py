'''
resticprofile file searching helpers
'''
from typing import List
from os import getcwd
from pathlib import Path, PosixPath, WindowsPath

# ==== If you guys need another location added to the default search paths, please make a PULL REQUEST ====
DEFAULT_CONFIGURATION_LOCATIONS_POSIX = [
    '/usr/local/etc/',
    '/usr/local/etc/restic/',
    '/usr/local/etc/resticprofile/',
    '/etc/',
    '/etc/restic/',
    '/etc/resticprofile/',
]

DEFAULT_CONFIGURATION_LOCATIONS_WINDOWS = [
    'c:\\restic\\',
    'c:\\resticprofile\\',
]

RESTIC_BINARY_POSIX = 'restic'
RESTIC_BINARY_WINDOWS = 'restic.exe'

DEFAULT_BINARY_LOCATIONS_POSIX = [
    '/usr/bin',
    '/usr/local/bin',
    '/opt/local/bin',
]

DEFAULT_BINARY_LOCATIONS_WINDOWS = [
    "c:\\ProgramData\\chocolatey\\bin\\",
    'c:\\restic\\',
    'c:\\resticprofile\\',
    'c:\\tools\\restic\\',
    'c:\\tools\\resticprofile\\',
]
# ========

def get_default_configuration_locations() -> List[str]:
    path = Path()
    if isinstance(path, PosixPath):
        return DEFAULT_CONFIGURATION_LOCATIONS_POSIX
    elif isinstance(path, WindowsPath):
        return DEFAULT_CONFIGURATION_LOCATIONS_WINDOWS

    return []


def get_default_binary_locations() -> List[str]:
    path = Path()
    if isinstance(path, PosixPath):
        return DEFAULT_BINARY_LOCATIONS_POSIX
    elif isinstance(path, WindowsPath):
        return DEFAULT_BINARY_LOCATIONS_WINDOWS

    return []

def get_restic_binary() -> str:
    path = Path()
    if isinstance(path, WindowsPath):
        return RESTIC_BINARY_WINDOWS

    return RESTIC_BINARY_POSIX

def find_configuration_file(configuration_file: str) -> str:
    '''
    Search for the file in the current directory, the home directory, and some pre-defined locations
    Returns None if the file was not found
    '''
    for filepath in list(
            map(
                lambda path: Path(path) / configuration_file,
                [getcwd(), str(Path().home())] + get_default_configuration_locations()
            )
        ):
        if filepath.is_file():
            return str(filepath)

    return None

def find_restic_binary() -> str:
    '''
    Search for restic binary in common locations (+ current directory and home directory)
    '''
    for filepath in list(
            map(
                lambda path: Path(path) / get_restic_binary(),
                [getcwd(), str(Path().home())] + get_default_binary_locations()
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
