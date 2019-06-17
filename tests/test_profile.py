import unittest
from src.lib.profile import Profile

class TestProfile(unittest.TestCase):

    def test_empty_configuration(self):
        profile = Profile('test')
        configuration_section = {}
        profile.set_global_configuration(configuration_section)
        self.assertEqual('test', profile.profile_name)

    def test_configuration_with_one_bool_true_flag(self):
        profile = Profile('test')
        configuration_section = { 'no-cache': True }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--no-cache" ])

    def test_configuration_with_one_bool_false_flag(self):
        profile = Profile('test')
        configuration_section = { 'no-cache': False }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [])

    def test_configuration_with_one_zero_int_flag(self):
        profile = Profile('test')
        configuration_section = { 'limit-upload': 0 }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--limit-upload 0" ])

    def test_configuration_with_one_positive_int_flag(self):
        profile = Profile('test')
        configuration_section = { 'limit-upload': 10 }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--limit-upload 10" ])

    def test_configuration_with_one_negative_int_flag(self):
        profile = Profile('test')
        configuration_section = { 'limit-upload': -10 }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--limit-upload -10" ])

    def test_configuration_with_repository(self):
        profile = Profile('test')
        configuration_section = { 'repository': "/backup" }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--repo '/backup'" ])

    def test_configuration_with_boolean_true_as_multiple_type_flag(self):
        profile = Profile('test')
        configuration_section = { 'verbose': True }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--verbose" ])

    def test_configuration_with_boolean_false_as_multiple_type_flag(self):
        profile = Profile('test')
        configuration_section = { 'verbose': False }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ ])

    def test_configuration_with_integer_as_multiple_type_flag(self):
        profile = Profile('test')
        configuration_section = { 'verbose': 1 }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--verbose 1" ])

    def test_configuration_with_wrong_type_as_multiple_type_flag(self):
        profile = Profile('test')
        configuration_section = { 'verbose': "toto" }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ ])

    def test_configuration_with_one_item_in_list_flag(self):
        profile = Profile('test')
        configuration_section = { 'option': "a=b" }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--option 'a=b'" ])

    def test_configuration_with_two_items_in_list_flag(self):
        profile = Profile('test')
        configuration_section = { 'option': [ "a=b", "b=c" ] }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(profile.get_global_flags(), [ "--option 'a=b'", "--option 'b=c'" ])

    def test_configuration_with_empty_repository(self):
        profile = Profile('test')
        configuration_section = { 'repository': '' }
        profile.set_global_configuration(configuration_section)
        self.assertFalse(profile.repository)

    def test_configuration_with_wrong_repository(self):
        profile = Profile('test')
        configuration_section = { 'repository': [ "one", "two" ] }
        profile.set_global_configuration(configuration_section)
        self.assertFalse(profile.repository)

    def test_configuration_with_valid_repository(self):
        profile = Profile('test')
        configuration_section = { 'repository': 'valid' }
        profile.set_global_configuration(configuration_section)
        self.assertEqual('valid', profile.repository)

    def test_configuration_with_quiet_flag(self):
        profile = Profile('test')
        configuration_section = { 'quiet': True }
        profile.set_global_configuration(configuration_section)
        self.assertTrue(profile.quiet)

    def test_configuration_with_verbose_flag(self):
        profile = Profile('test')
        configuration_section = { 'verbose': 2 }
        profile.set_global_configuration(configuration_section)
        self.assertEqual(2, profile.verbose)
