import unittest
from resticprofile.config import Config
from resticprofile.profile import Profile

class MockFileSearch:
    pass

class TestProfile(unittest.TestCase):

    def new_profile(self, configuration: dict):
        return Profile(Config(configuration, MockFileSearch()), 'test')

    def test_no_configuration(self):
        configuration = {}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual('test', profile.profile_name)

    def test_empty_configuration(self):
        configuration = {'test': {}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual('test', profile.profile_name)

    def test_configuration_with_one_bool_true_flag(self):
        configuration = {'test': {'no-cache': True}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--no-cache"])

    def test_configuration_with_one_bool_false_flag(self):
        configuration = {'test': {'no-cache': False}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), [])

    def test_configuration_with_one_zero_int_flag(self):
        configuration = {'test': {'limit-upload': 0}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--limit-upload 0"])

    def test_configuration_with_one_positive_int_flag(self):
        configuration = {'test': {'limit-upload': 10}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--limit-upload 10"])

    def test_configuration_with_one_negative_int_flag(self):
        configuration = {'test': {'limit-upload': -10}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--limit-upload -10"])

    def test_configuration_with_repository(self):
        configuration = {'test': {'repository': "/backup"}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--repo '/backup'"])

    def test_configuration_with_boolean_true_as_multiple_type_flag(self):
        configuration = {'test': {'verbose': True}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--verbose"])

    def test_configuration_with_boolean_false_as_multiple_type_flag(self):
        configuration = {'test': {'verbose': False}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), [])

    def test_configuration_with_integer_as_multiple_type_flag(self):
        configuration = {'test': {'verbose': 1}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--verbose 1"])

    def test_configuration_with_wrong_type_as_multiple_type_flag(self):
        configuration = {'test': {'verbose': "toto"}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), [])

    def test_configuration_with_one_item_in_list_flag(self):
        configuration = {'test': {'option': "a=b"}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--option 'a=b'"])

    def test_configuration_with_two_items_in_list_flag(self):
        configuration = {'test': {'option': ["a=b", "b=c"]}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), [
                         "--option 'a=b'", "--option 'b=c'"])

    def test_configuration_with_empty_repository(self):
        configuration = {'test': {'repository': ''}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertFalse(profile.repository)

    def test_configuration_with_wrong_repository(self):
        configuration = {'test': {'repository': ["one", "two"]}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertFalse(profile.repository)

    def test_configuration_with_valid_repository(self):
        configuration = {'test': {'repository': 'valid'}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual('valid', profile.repository)

    def test_configuration_with_quiet_flag(self):
        configuration = {'test': {'quiet': True}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertTrue(profile.quiet)

    def test_configuration_with_verbose_flag(self):
        configuration = {'test': {'verbose': 2}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(2, profile.verbose)

    def test_configuration_with_inherited_verbose_flag(self):
        configuration = {'parent': {'verbose': True}, 'test': {'inherit': 'parent'}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(True, profile.verbose)
        self.assertEqual(profile.inherit, 'parent')

    def test_configuration_with_inherited_repository(self):
        configuration = {'parent': {'repository': "/backup"}, 'test': {'inherit': 'parent'}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertEqual(profile.get_global_flags(), ["--repo '/backup'"])
