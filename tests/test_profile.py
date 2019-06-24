import unittest
from pathlib import Path

from resticprofile.config import Config
from resticprofile.profile import Profile
from mock_filesearch import MockFileSearch

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
        self.assertEqual(profile.get_global_flags(), ["--repo \"/backup\""])

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
        self.assertEqual(profile.get_global_flags(), ["--option \"a=b\""])

    def test_configuration_with_two_items_in_list_flag(self):
        configuration = {'test': {'option': ["a=b", "b=c"]}}
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        self.assertCountEqual(profile.get_global_flags(), ["--option \"a=b\"", "--option \"b=c\""])

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
        self.assertEqual(profile.get_global_flags(), ["--repo \"/backup\""])

    def test_empty_command_configuration(self):
        configuration = {
            'test': {
                'backup': {}
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_command_flags('backup'), [])

    def test_empty_section_command_configuration(self):
        configuration = {
            'test': {
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_command_flags('backup'), [])

    def test_empty_configuration_command_configuration(self):
        configuration = {}
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_command_flags('backup'), [])


    def test_command_configuration(self):
        configuration = {
            'test': {
                'backup': {
                    'source': 'folder'
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_global_flags(), [])
        self.assertEqual(profile.get_command_flags('backup'), ["\"folder\""])


    def test_inherited_command_configuration(self):
        configuration = {
            'parent': {
                'backup': {
                    'tag': 'parent'
                }
            },
            'test': {
                'inherit': 'parent',
                'backup': {
                    'source': 'folder'
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_global_flags(), [])
        self.assertCountEqual(profile.get_command_flags('backup'), ["--tag \"parent\"", "\"folder\""])


    def test_overridden_command_configuration(self):
        configuration = {
            'parent': {
                'backup': {
                    'tag': 'parent',
                    'source': 'folder1'
                }
            },
            'test': {
                'inherit': 'parent',
                'backup': {
                    'source': 'folder2'
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_global_flags(), [])
        self.assertCountEqual(profile.get_command_flags('backup'), ["--tag \"parent\"", "\"folder1\"", "\"folder2\""])

    def test_twice_inherited_command_configuration(self):
        configuration = {
            'grand-parent': {
                'backup': {
                    'tag': 'grand-parent'
                }
            },
            'parent': {
                'inherit': 'grand-parent',
                'backup': {}
            },
            'test': {
                'inherit': 'parent',
                'backup': {
                    'source': 'folder'
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_global_flags(), [])
        self.assertCountEqual(profile.get_command_flags('backup'), ["--tag \"grand-parent\"", "\"folder\""])

    def test_twice_overridden_command_configuration(self):
        configuration = {
            'grand-parent': {
                'backup': {
                    'tag': 'grand-parent'
                }
            },
            'parent': {
                'inherit': 'grand-parent',
                'backup': {
                    'tag': 'parent'
                }
            },
            'test': {
                'inherit': 'parent',
                'backup': {
                    'source': 'folder'
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_global_flags(), [])
        self.assertCountEqual(profile.get_command_flags('backup'), ["--tag \"parent\"", "\"folder\""])


    def test_removing_duplicates_command_configuration(self):
        configuration = {
            'parent': {
                'backup': {
                    'source': 'folder'
                }
            },
            'test': {
                'inherit': 'parent',
                'backup': {
                    'source': 'folder'
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_command_configuration('backup')

        self.assertEqual(profile.get_global_flags(), [])
        self.assertCountEqual(profile.get_command_flags('backup'), ["\"folder\""])

    def test_direct_false_initialize_flag(self):
        configuration = {
            'test': {
                'initialize': False,
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()

        self.assertFalse(profile.initialize)

    def test_direct_true_initialize_flag(self):
        configuration = {
            'test': {
                'initialize': True,
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()

        self.assertTrue(profile.initialize)

    def test_direct_command_false_initialize_flag(self):
        configuration = {
            'test': {
                'initialize': True,
                'backup': {
                    'initialize': False,
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')

        self.assertFalse(profile.initialize)

    def test_direct_command_true_initialize_flag(self):
        configuration = {
            'test': {
                'initialize': False,
                'backup': {
                    'initialize': True,
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')

        self.assertTrue(profile.initialize)


    def test_inherited_false_initialize_flag(self):
        configuration = {
            'parent': {
                'initialize': False,
            },
            'test': {
                'inherit': 'parent',
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()

        self.assertFalse(profile.initialize)

    def test_inherited_true_initialize_flag(self):
        configuration = {
            'parent': {
                'initialize': True,
            },
            'test': {
                'inherit': 'parent',
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()

        self.assertTrue(profile.initialize)

    def test_inherited_command_false_initialize_flag(self):
        configuration = {
            'parent': {
                'initialize': True,
                'backup': {
                    'initialize': False,
                }
            },
            'test': {
                'inherit': 'parent',
                'backup': {},
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')

        self.assertFalse(profile.initialize)

    def test_inherited_command_true_initialize_flag(self):
        configuration = {
            'parent': {
                'initialize': False,
                'backup': {
                    'initialize': True,
                }
            },
            'test': {
                'inherit': 'parent',
                'backup': {},
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')

        self.assertTrue(profile.initialize)

    def test_backup_profile_is_loading_retention_flags_with_path(self):
        configuration = {
            'test': {
                'backup': {
                    'source': '/source',
                },
                'retention': {
                    'keep-last': 3
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')
        profile.set_retention_configuration()
        retention_flags = profile.get_retention_flags()
        self.assertCountEqual(["--keep-last 3", "--path \"{}\"".format(str(Path('/source')))], retention_flags)

    def test_backup_profile_is_loading_retention_path_flags_without_path(self):
        configuration = {
            'test': {
                'backup': {
                    'source': '/source',
                },
                'retention': {
                    'keep-last': 3,
                    'path': '/other'
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')
        profile.set_retention_configuration()
        retention_flags = profile.get_retention_flags()
        self.assertCountEqual(["--keep-last 3", "--path \"{}\"".format(str(Path('/other')))], retention_flags)

    def test_backup_profile_is_loading_retention_path_flags_with_empty_path(self):
        configuration = {
            'test': {
                'backup': {
                    'source': '/source',
                },
                'retention': {
                    'keep-last': 3,
                    'path': []
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')
        profile.set_retention_configuration()
        retention_flags = profile.get_retention_flags()
        self.assertEqual(["--keep-last 3"], retention_flags)

    def test_backup_profile_is_not_loading_retention_tag_flags_from_backup(self):
        configuration = {
            'test': {
                'backup': {
                    'tag': 'backup-tag',
                },
                'retention': {
                    'keep-last': 3,
                }
            }
        }
        profile = self.new_profile(configuration)
        profile.set_common_configuration()
        profile.set_command_configuration('backup')
        profile.set_retention_configuration()
        retention_flags = profile.get_retention_flags()
        self.assertEqual(["--keep-last 3"], retention_flags)
