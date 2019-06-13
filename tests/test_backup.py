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
        self.assertEqual(backup.flags, [ "--no-cache" ])

    def test_configuration_with_one_bool_false_flag(self):
        backup = Backup('test')
        configuration_section = { 'no-cache': False }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.flags, [])

    def test_configuration_with_repository(self):
        backup = Backup('test')
        configuration_section = { 'repository': "/backup" }
        backup.set_global_configuration(configuration_section)
        self.assertEqual(backup.flags, [ "--repo=/backup" ])
