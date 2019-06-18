import unittest
from src.lib.flag import Flag

class TestProfile(unittest.TestCase):

    def test_can_get_string_value(self):
        flag = Flag('key', 'value', 'str')
        argument = flag.get_flags()
        self.assertEqual(argument, ["--key 'value'"])

    def test_can_get_empty_string_value(self):
        flag = Flag('key', '', 'str')
        argument = flag.get_flags()
        self.assertEqual(argument, ["--key ''"])

    def test_can_get_true_boolean_value(self):
        flag = Flag('key', True, 'bool')
        argument = flag.get_flags()
        self.assertEqual(argument, ["--key"])

    def test_can_get_false_boolean_value(self):
        flag = Flag('key', False, 'bool')
        argument = flag.get_flags()
        self.assertEqual(argument, [])

    def test_can_get_zero_int_value(self):
        flag = Flag('key', 0, 'int')
        argument = flag.get_flags()
        self.assertEqual(argument, ["--key 0"])

    def test_can_get_positive_int_value(self):
        flag = Flag('key', 1, 'int')
        argument = flag.get_flags()
        self.assertEqual(argument, ["--key 1"])

    def test_can_get_negative_int_value(self):
        flag = Flag('key', -1, 'int')
        argument = flag.get_flags()
        self.assertEqual(argument, ["--key -1"])

    def test_can_get_string_values(self):
        flag = Flag('key', ["1", "2"], 'str')
        argument = flag.get_flags()
        self.assertEqual(argument, ["--key '1'", "--key '2'"])
