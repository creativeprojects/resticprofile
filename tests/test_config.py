import unittest
from src.lib.config import validate_configuration_option

class TestConfig(unittest.TestCase):

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
        self.assertEqual(result, {'key': 'key', 'value': True, 'type': 'bool'})

    def test_boolean_false(self):
        definition = { 'key': {'type': 'bool'}}
        result = validate_configuration_option(definition, 'key', False)
        self.assertEqual(result, {'key': 'key', 'value': False, 'type': 'bool'})

    def test_no_boolean(self):
        definition = { 'key': {'type': 'bool'}}
        result = validate_configuration_option(definition, 'key', 0)
        self.assertFalse(result)

    def test_positive_integer(self):
        definition = { 'key': {'type': 'int'}}
        result = validate_configuration_option(definition, 'key', 10)
        self.assertEqual(result, {'key': 'key', 'value': 10, 'type': 'int'})

    def test_negative_integer(self):
        definition = { 'key': {'type': 'int'}}
        result = validate_configuration_option(definition, 'key', -10)
        self.assertEqual(result, {'key': 'key', 'value': -10, 'type': 'int'})

    def test_no_integer(self):
        definition = { 'key': {'type': 'int'}}
        result = validate_configuration_option(definition, 'key', '0')
        self.assertFalse(result)

    def test_string(self):
        definition = { 'key': {'type': 'str'}}
        result = validate_configuration_option(definition, 'key', 'something')
        self.assertEqual(result, {'key': 'key', 'value': 'something', 'type': 'str'})

    def test_no_string(self):
        definition = { 'key': {'type': 'str'}}
        result = validate_configuration_option(definition, 'key', 0)
        self.assertFalse(result)

    def test_bool_or_int_with_false(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', False)
        self.assertEqual(result, {'key': 'key', 'value': False, 'type': 'bool'})

    def test_bool_or_int_with_true(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', True)
        self.assertEqual(result, {'key': 'key', 'value': True, 'type': 'bool'})

    def test_bool_or_int_with_positive_int(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', 1)
        self.assertEqual(result, {'key': 'key', 'value': 1, 'type': 'int'})

    def test_bool_or_int_with_negative_int(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', -1)
        self.assertEqual(result, {'key': 'key', 'value': -1, 'type': 'int'})

    def test_bool_or_int_with_zero(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', 0)
        self.assertEqual(result, {'key': 'key', 'value': 0, 'type': 'int'})

    def test_bool_or_int_with_string(self):
        definition = { 'key': {'type': ['bool', 'int'] } }
        result = validate_configuration_option(definition, 'key', "0")
        self.assertFalse(result)

    def test_list_of_strings(self):
        definition = { 'key': {'type': 'str', 'list': True } }
        result = validate_configuration_option(definition, 'key', ["1", "2"])
        self.assertEqual(result, {'key': 'key', 'value': ["1", "2"], 'type': 'str'})

    def test_wrong_list_of_strings(self):
        definition = { 'key': {'type': 'str', 'list': True } }
        result = validate_configuration_option(definition, 'key', ["1", 2])
        self.assertFalse(result)

    def test_flag_with_different_name_betwen_configuration_file_and_restic_command_line(self):
        definition = { 'key': {'type': 'str', 'flag': 'other-key' } }
        result = validate_configuration_option(definition, 'key', "value")
        self.assertEqual(result, {'key': 'other-key', 'value': "value", 'type': 'str'})
