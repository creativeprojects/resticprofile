import unittest

from resticprofile.groups import Groups

class TestProfile(unittest.TestCase):

    def test_with_no_group_section(self):
        configuration = {}
        groups = Groups(configuration)
        self.assertFalse(groups.exists('test'))

    def test_with_empty_group_section(self):
        configuration = {
            'groups': {}
        }
        groups = Groups(configuration)
        self.assertFalse(groups.exists('test'))

    def test_with_wrong_group_profile(self):
        configuration = {
            'groups': {
                'test': 'profile'
            }
        }
        groups = Groups(configuration)
        self.assertFalse(groups.exists('test'))

    def test_with_group_with_no_profile(self):
        configuration = {
            'groups': {
                'test': []
            }
        }
        groups = Groups(configuration)
        self.assertFalse(groups.exists('test'))

    def test_with_group_with_one_profile(self):
        configuration = {
            'groups': {
                'test': ['profile']
            }
        }
        groups = Groups(configuration)
        self.assertTrue(groups.exists('test'))
        self.assertEqual(['profile'], groups.get_profiles('test'))

    def test_with_group_with_two_profiles(self):
        configuration = {
            'groups': {
                'full-backup': ['dev', 'src']
            }
        }
        groups = Groups(configuration)
        self.assertTrue(groups.exists('full-backup'))
        self.assertCountEqual(['dev', 'src'], groups.get_profiles('full-backup'))
