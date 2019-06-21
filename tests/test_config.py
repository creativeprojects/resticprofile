import unittest
from resticprofile.config import Config, DEFAULTS
from resticprofile.flag import Flag
from resticprofile.ionice import IONice

class MockFileSearch:
    pass

class TestConfig(unittest.TestCase):

    def new_config(self, configuration: dict):
        return Config(configuration, MockFileSearch())

    def assertFlag(self, flag, key, value, type_value):
        self.assertEqual(key, flag.key, "Expected key was {} but found {}".format(key, flag.key))
        self.assertEqual(value, flag.value, "Expected value was {} but found {}".format(value, flag.value))
        self.assertEqual(type_value, flag.type, "Expected type was {} but found {}".format(type_value, flag.type))

    def test_unknown_definition(self):
        definition = {}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 'value')
        self.assertFalse(result)

    def test_unknown_type(self):
        definition = {'key': {}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 'value')
        self.assertFalse(result)

    def test_unknown_type_should_raise_excepton(self):
        definition = {'key': {'type': 'unknown'}}
        with self.assertRaises(Exception) as context:
            self.new_config({}).validate_configuration_option(definition, 'key', 'value')
        self.assertEqual("Unknown type 'unknown'", str(context.exception))

    def test_boolean_true(self):
        definition = {'key': {'type': 'bool'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', True)
        self.assertFlag(result, 'key', True, 'bool')

    def test_boolean_false(self):
        definition = {'key': {'type': 'bool'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', False)
        self.assertFlag(result, 'key', False, 'bool')

    def test_no_boolean(self):
        definition = {'key': {'type': 'bool'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 0)
        self.assertFalse(result)

    def test_positive_integer(self):
        definition = {'key': {'type': 'int'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 10)
        self.assertFlag(result, 'key', 10, 'int')

    def test_negative_integer(self):
        definition = {'key': {'type': 'int'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', -10)
        self.assertFlag(result, 'key', -10, 'int')

    def test_no_integer(self):
        definition = {'key': {'type': 'int'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', '0')
        self.assertFalse(result)

    def test_string(self):
        definition = {'key': {'type': 'str'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 'something')
        self.assertFlag(result, 'key', 'something', 'str')

    def test_no_string(self):
        definition = {'key': {'type': 'str'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 0)
        self.assertFalse(result)

    def test_bool_or_int_with_false(self):
        definition = {'key': {'type': ['bool', 'int']}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', False)
        self.assertFlag(result, 'key', False, 'bool')

    def test_bool_or_int_with_true(self):
        definition = {'key': {'type': ['bool', 'int']}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', True)
        self.assertFlag(result, 'key', True, 'bool')

    def test_bool_or_int_with_positive_int(self):
        definition = {'key': {'type': ['bool', 'int']}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 1)
        self.assertFlag(result, 'key', 1, 'int')

    def test_bool_or_int_with_negative_int(self):
        definition = {'key': {'type': ['bool', 'int']}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', -1)
        self.assertFlag(result, 'key', -1, 'int')

    def test_bool_or_int_with_zero(self):
        definition = {'key': {'type': ['bool', 'int']}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', 0)
        self.assertFlag(result, 'key', 0, 'int')

    def test_bool_or_int_with_string(self):
        definition = {'key': {'type': ['bool', 'int']}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', "0")
        self.assertFalse(result)

    def test_list_of_strings(self):
        definition = {'key': {'type': 'str', 'list': True}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', ["1", "2"])
        self.assertFlag(result, 'key', ["1", "2"], 'str')

    def test_wrong_list_of_strings(self):
        definition = {'key': {'type': 'str', 'list': True}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', ["1", 2])
        self.assertFalse(result)

    def test_flag_with_different_name_betwen_configuration_file_and_restic_command_line(self):
        definition = {'key': {'type': 'str', 'flag': 'other-key'}}
        result = self.new_config({}).validate_configuration_option(definition, 'key', "value")
        self.assertFlag(result, 'other-key', 'value', 'str')

    # nice
    def test_nice_zero(self):
        configuration = {
            'global': {'nice': 0}
        }
        nice = self.new_config(configuration).get_nice()
        self.assertEqual(0, nice.niceness)

    def test_nice_positive(self):
        configuration = {
            'global': {'nice': 1}
        }
        nice = self.new_config(configuration).get_nice()
        self.assertEqual(1, nice.niceness)

    def test_nice_negative(self):
        configuration = {
            'global': {'nice': -1}
        }
        nice = self.new_config(configuration).get_nice()
        self.assertEqual(-1, nice.niceness)

    def test_nice_false(self):
        configuration = {
            'global': {'nice': False}
        }
        nice = self.new_config(configuration).get_nice()
        self.assertEqual(None, nice)

    def test_without_nice(self):
        configuration = {'global': {}}
        nice = self.new_config(configuration).get_nice()
        self.assertEqual(None, nice)

    # ionice
    def test_without_ionice(self):
        configuration = {}
        ionice = self.new_config(configuration).get_ionice()
        self.assertEqual(None, ionice)

    def test_no_ionice(self):
        configuration = {
            'global': {'ionice': False}
        }
        ionice = self.new_config(configuration).get_ionice()
        self.assertEqual(None, ionice)

    def test_empty_ionice(self):
        configuration = {
            'global': {'ionice': True}
        }
        ionice = self.new_config(configuration).get_ionice()
        self.assertIsInstance(ionice, IONice)

    # default-command
    def test_no_default_command(self):
        configuration = {}
        default_command = self.new_config(configuration).get_default_command()
        self.assertEqual(DEFAULTS['default_command'], default_command)

    def test_wrong_default_command(self):
        configuration = {
            'global': {'default-command': False}
        }
        default_command = self.new_config(configuration).get_default_command()
        self.assertEqual(DEFAULTS['default_command'], default_command)

    def test_default_command(self):
        configuration = {
            'global': {'default-command': 'test_test'}
        }
        default_command = self.new_config(configuration).get_default_command()
        self.assertEqual('test_test', default_command)

    # initialize
    def test_no_initialize(self):
        configuration = {}
        initialize = self.new_config(configuration).get_initialize()
        self.assertEqual(DEFAULTS['initialize'], initialize)

    def test_wrong_initialize(self):
        configuration = {
            'global': {'initialize': 0}
        }
        initialize = self.new_config(configuration).get_initialize()
        self.assertEqual(DEFAULTS['initialize'], initialize)

    def test_initialize(self):
        configuration = {
            'global': {'initialize': True}
        }
        initialize = self.new_config(configuration).get_initialize()
        self.assertTrue(initialize)


    def test_loading_no_common_options(self):
        configuration = {}
        options = self.new_config(configuration).get_options_for_section('test')
        self.assertEqual([], options)

    def test_loading_empty_common_options(self):
        configuration = {
            'test': {}
        }
        options = self.new_config(configuration).get_options_for_section('test')
        self.assertEqual([], options)

    def test_loading_simple_common_options(self):
        configuration = {
            'test': {'repository': '/backup'}
        }
        options = self.new_config(configuration).get_options_for_section('test')
        self.assertEqual(len(options), 1)
        self.assertIsInstance(options[0], Flag)

    def test_loading_unknown_inherited_common_options(self):
        configuration = {
            'test': {'inherit': 'parent'}
        }
        options = self.new_config(configuration).get_options_for_section('test')
        # at that stage, config is keeping the inherit flag because it will be needed up the line
        self.assertEqual(len(options), 1)
        self.assertIsInstance(options[0], Flag)

    def test_loading_inherited_common_options(self):
        configuration = {
            'parent': {'repository': '/backup'},
            'test': {'inherit': 'parent'}
        }
        options = self.new_config(configuration).get_options_for_section('test')
        # at that stage, config is keeping the inherit flag because it will be needed up the line
        self.assertEqual(len(options), 2)
        self.assertIsInstance(options[0], Flag)
        self.assertIsInstance(options[1], Flag)

    def test_loading_twice_inherited_common_options(self):
        configuration = {
            'grand-parent': {'no-cache': True},
            'parent': {'inherit': 'grand-parent', 'repository': '/backup'},
            'test': {'inherit': 'parent'}
        }
        options = self.new_config(configuration).get_options_for_section('test')
        # at that stage, config is keeping both the inherit flags because they will be needed up the line
        self.assertEqual(len(options), 4)
        self.assertIsInstance(options[0], Flag)
        self.assertIsInstance(options[1], Flag)
        self.assertIsInstance(options[2], Flag)
        self.assertIsInstance(options[3], Flag)

    def test_loading_command_section(self):
        configuration = {
            'profile': {
                'backup': {
                    'source': 'folder'
                }
            },
        }
        options = self.new_config(configuration).get_options_for_section('profile', 'backup')
        self.assertEqual(1, len(options))
        self.assertIsInstance(options[0], Flag)
        self.assertEqual('source', options[0].key)


    def test_can_get_initialize_flag_from_profile(self):
        configuration = {
            'profile': {
                'initialize': True
            },
        }
        options = self.new_config(configuration).get_options_for_section('profile')
        self.assertEqual(1, len(options))
        self.assertIsInstance(options[0], Flag)

    def test_can_get_initialize_flag_from_command_profile(self):
        configuration = {
            'profile': {
                'backup': {
                    'initialize': True
                }
            },
        }
        options = self.new_config(configuration).get_options_for_section('profile', 'backup')
        self.assertEqual(1, len(options))
        self.assertIsInstance(options[0], Flag)

    def test_can_get_initialize_flag_from_inherited_profile(self):
        configuration = {
            'parent': {
                'initialize': True
            },
            'profile': {
                'inherit': 'parent',
            },
        }
        options = self.new_config(configuration).get_options_for_section('profile')
        self.assertEqual(2, len(options))
        self.assertIsInstance(options[0], Flag)
        self.assertIsInstance(options[1], Flag)

    def test_can_get_initialize_flag_from_inherited_command_profile(self):
        configuration = {
            'parent': {
                'backup': {
                    'initialize': True
                }
            },
            'profile': {
                'inherit': 'parent',
                'backup': {}
            },
        }
        options = self.new_config(configuration).get_options_for_section('profile', 'backup')
        self.assertEqual(2, len(options))
        self.assertIsInstance(options[0], Flag)
        self.assertIsInstance(options[1], Flag)
