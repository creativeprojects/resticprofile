import unittest
from src.lib.backup import Backup

class TestBackup(unittest.TestCase):

    def test_empty_configuration(self):
        backup = Backup('test')
        configuration_section = {}
        backup.set_global_configuration(configuration_section)
        self.assertEqual('test', backup.profile_name)

    def test_configuration_with_one_bool_true_flag(self):
        backup = Backup('test')
        configuration_section = { 'no-cache': True }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ "--no-cache" ])

    def test_configuration_with_one_bool_false_flag(self):
        backup = Backup('test')
        configuration_section = { 'no-cache': False }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [])

    def test_configuration_with_repository(self):
        backup = Backup('test')
        configuration_section = { 'repository': "/backup" }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ "--repo /backup" ])

    def test_configuration_with_boolean_true_as_multiple_type_flag(self):
        backup = Backup('test')
        configuration_section = { 'verbose': True }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ "--verbose" ])

    def test_configuration_with_boolean_false_as_multiple_type_flag(self):
        backup = Backup('test')
        configuration_section = { 'verbose': False }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ ])

    def test_configuration_with_integer_as_multiple_type_flag(self):
        backup = Backup('test')
        configuration_section = { 'verbose': 1 }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ "--verbose 1" ])

    def test_configuration_with_wrong_type_as_multiple_type_flag(self):
        backup = Backup('test')
        configuration_section = { 'verbose': "toto" }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ ])

    def test_configuration_with_one_item_in_list_flag(self):
        backup = Backup('test')
        configuration_section = { 'option': "a=b" }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ "--option a=b" ])

    def test_configuration_with_two_items_in_list_flag(self):
        backup = Backup('test')
        configuration_section = { 'option': [ "a=b", "b=c" ] }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.global_flags, [ "--option a=b", "--option b=c" ])

    def test_configuration_with_empty_repository(self):
        backup = Backup('test')
        configuration_section = { 'repository': '' }
        backup.set_global_configuration(configuration_section)
        self.assertFalse(backup.repository)

    def test_configuration_with_wrong_repository(self):
        backup = Backup('test')
        configuration_section = { 'repository': [ "one", "two" ] }
        backup.set_global_configuration(configuration_section)
        self.assertFalse(backup.repository)

    def test_configuration_with_valid_repository(self):
        backup = Backup('test')
        configuration_section = { 'repository': 'valid' }
        backup.set_global_configuration(configuration_section)
        self.assertEqual('valid', backup.repository)

    def test_configuration_with_quiet_flag(self):
        backup = Backup('test')
        configuration_section = { 'quiet': True }
        backup.set_global_configuration(configuration_section)
        self.assertTrue(backup.quiet)

    def test_configuration_with_verbose_flag(self):
        backup = Backup('test')
        configuration_section = { 'verbose': 2 }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(2, backup.verbose)
