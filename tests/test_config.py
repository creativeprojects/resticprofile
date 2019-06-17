import unittest
from src.lib.config import validate_configuration_option

class TestConfig(unittest.TestCase):

    def assertFlag(self, flag, key, value, type_value):
        self.assertEqual(key, flag.key, "Expected key was {} but found {}".format(key, flag.key))
        self.assertEqual(value, flag.value, "Expected value was {} but found {}".format(value, flag.value))
        self.assertEqual(type_value, flag.type, "Expected type was {} but found {}".format(type_value, flag.type))

    def test_unknown_definition(self):
        definition = {}
        result = validate_configuration_option(definition, 'key', 'value')
        self.assertFalse(result)

    def test_unknown_type(self):
        definition = { 'key': {} }
        result = validate_configuration_option(definition, 'key', 'value')
        self.assertFalse(result)

    def test_unknown_type_should_raise_excepton(self):
        definition = { 'key': {'type': 'unknown'}}
        with self.assertRaises(Exception) as context:
            validate_configuration_option(definition, 'key', 'value')
        self.assertEqual("Unknown type 'unknown'", str(context.exception))

    def test_boolean_true(self):
        definition = { 'key': {'type': 'bool'}}
        result = validate_configuration_option(definition, 'key', True)
        self.assertFlag(result, 'key', True, 'bool')

    def test_boolean_false(self):
        definition = { 'key': {'type': 'bool'}}
        result = validate_configuration_option(definition, 'key', False)
        self.assertFlag(result, 'key', False, 'bool')

    def test_no_boolean(self):
        definition = { 'key': {'type': 'bool'}}
        result = validate_configuration_option(definition, 'key', 0)
        self.assertFalse(result)

    def test_positive_integer(self):
        definition = { 'key': {'type': 'int'}}
        result = validate_configuration_option(definition, 'key', 10)
        self.assertFlag(result, 'key', 10, 'int')

    def test_negative_integer(self):
        definition = { 'key': {'type': 'int'}}
        result = validate_configuration_option(definition, 'key', -10)
        self.assertFlag(result, 'key', -10, 'int')

    def test_no_integer(self):
        definition = { 'key': {'type': 'int'}}
        result = validate_configuration_option(definition, 'key', '0')
        self.assertFalse(result)

    def test_string(self):
        definition = { 'key': {'type': 'str'}}
        result = validate_configuration_option(definition, 'key', 'something')
        self.assertFlag(result, 'key', 'something', 'str')

    def test_no_string(self):
        definition = { 'key': {'type': 'str'}}
        result = validate_configuration_option(definition, 'key', 0)
        self.assertFalse(result)

    def test_bool_or_int_with_false(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', False)
        self.assertFlag(result, 'key', False, 'bool')

    def test_bool_or_int_with_true(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', True)
        self.assertFlag(result, 'key', True, 'bool')

    def test_bool_or_int_with_positive_int(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', 1)
        self.assertFlag(result, 'key', 1, 'int')

    def test_bool_or_int_with_negative_int(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', -1)
        self.assertFlag(result, 'key', -1, 'int')

    def test_bool_or_int_with_zero(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', 0)
        self.assertFlag(result, 'key', 0, 'int')

    def test_bool_or_int_with_string(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', "0")
        self.assertFalse(result)

    def test_list_of_strings(self):
        definition = { 'key': {'type': 'str', 'list': True } }
        result = validate_configuration_option(definition, 'key', ["1", "2"])
        self.assertFlag(result, 'key', ["1", "2"], 'str')

    def test_wrong_list_of_strings(self):
        definition = { 'key': {'type': 'str', 'list': True } }
        result = validate_configuration_option(definition, 'key', ["1", 2])
        self.assertFalse(result)

    def test_flag_with_different_name_betwen_configuration_file_and_restic_command_line(self):
        definition = { 'key': {'type': 'str', 'flag': 'other-key' } }
        result = validate_configuration_option(definition, 'key', "value")
        self.assertFlag(result, 'other-key', 'value', 'str')
