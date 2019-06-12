import unittest
from src.lib.help import get_options_help

mock_arguments_definition = {
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

    def test_can_load_arguments_definition(self):
        args_help = get_options_help(mock_arguments_definition)
        self.assertListEqual(args_help, ['[-h|--help]', '[-n|--name <profile_name>]'])
