import unittest
from src.lib.help import get_options_help

mock_ARGUMENTS_DEFINITION = {
    'help': {
        'short': 'h',
        'long': 'help',
        'argument': False,
    },
    'error': {
    },
    'name': {
        'short': 'n',
        'long': 'name',
        'argument': True,
        'argument_name': 'profile_name'
    },
}

class TestHelp(unittest.TestCase):

    def test_can_load_ARGUMENTS_DEFINITION(self):
        args_help = get_options_help(mock_ARGUMENTS_DEFINITION)
        self.assertListEqual(args_help, ['[-h|--help]', '[-n|--name <profile_name>]'])
