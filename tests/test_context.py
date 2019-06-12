import unittest
from src.lib.config import defaults
from src.lib.context import Context
from src.lib.ionice import IONice

class TestContext(unittest.TestCase):

    # nice
    def test_nice_zero(self):
        context = Context()
        global_section = {
            'nice': 0
        }
        context.set_global_context(global_section)
        self.assertEqual(0, context.nice.niceness)

    def test_nice_positive(self):
        context = Context()
        global_section = {
            'nice': 1
        }
        context.set_global_context(global_section)
        self.assertEqual(1, context.nice.niceness)

    def test_nice_negative(self):
        context = Context()
        global_section = {
            'nice': -1
        }
        context.set_global_context(global_section)
        self.assertEqual(-1, context.nice.niceness)

    def test_nice_false(self):
        context = Context()
        global_section = {
            'nice': False
        }
        context.set_global_context(global_section)
        self.assertEqual(None, context.nice)

    def test_without_nice(self):
        context = Context()
        global_section = {}
        context.set_global_context(global_section)
        self.assertEqual(None, context.nice)

    # ionice
    def test_without_ionice(self):
        context = Context()
        global_section = {}
        context.set_global_context(global_section)
        self.assertEqual(None, context.ionice)

    def test_no_ionice(self):
        context = Context()
        global_section = {
            'ionice': False
        }
        context.set_global_context(global_section)
        self.assertEqual(None, context.ionice)

    def test_empty_ionice(self):
        context = Context()
        global_section = {
            'ionice': True
        }
        context.set_global_context(global_section)
        self.assertIsInstance(context.ionice, IONice)

    # default-command
    def test_no_default_command(self):
        context = Context()
        global_section = {}
        context.set_global_context(global_section)
        self.assertEqual(defaults['default_command'], context.default_command)

    def test_wrong_default_command(self):
        context = Context()
        global_section = {
            'default-command': False
        }
        context.set_global_context(global_section)
        self.assertEqual(defaults['default_command'], context.default_command)

    def test_default_command(self):
        context = Context()
        global_section = {
            'default-command': 'test_test'
        }
        context.set_global_context(global_section)
        self.assertEqual('test_test', context.default_command)

    # initialize
    def test_no_initialize(self):
        context = Context()
        global_section = {}
        context.set_global_context(global_section)
        self.assertEqual(defaults['initialize'], context.initialize)

    def test_wrong_initialize(self):
        context = Context()
        global_section = {
            'initialize': 0
        }
        context.set_global_context(global_section)
        self.assertEqual(defaults['initialize'], context.initialize)

    def test_initialize(self):
        context = Context()
        global_section = {
            'initialize': True
        }
        context.set_global_context(global_section)
        self.assertTrue(context.initialize)
