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
