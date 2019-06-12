import unittest
from src.lib.ionice import IONice

class TestIONice(unittest.TestCase):

    def test_can_get_command_with_default_value(self):
        ionice = IONice()
        command = ionice.get_command('Linux')
        self.assertEqual(command, "ionice -c 2 -n 4")

    def test_can_get_command_for_none_class(self):
        ionice = IONice(0)
        command = ionice.get_command('Linux')
        self.assertEqual(command, "ionice -c 0")

    def test_can_get_command_for_realtime_class(self):
        ionice = IONice(1)
        command = ionice.get_command('Linux')
        self.assertEqual(command, "ionice -c 1 -n 4")

    def test_can_get_command_for_besteffort_class(self):
        ionice = IONice(2)
        command = ionice.get_command('Linux')
        self.assertEqual(command, "ionice -c 2 -n 4")

    def test_can_get_command_for_idle_class(self):
        ionice = IONice(3)
        command = ionice.get_command('Linux')
        self.assertEqual(command, "ionice -c 3")

    def test_can_get_command_with_ignore_failure(self):
        ionice = IONice(io_class=3, ignore_failure=True)
        command = ionice.get_command('Linux')
        self.assertEqual(command, "ionice -c 3 -t")

    def test_can_get_command_will_not_throw_exception_and_return_empty_string(self):
        ionice = IONice(ignore_failure=True)
        command = ionice.get_command('Windows')
        self.assertEqual(command, "")

if __name__ == '__main__':
    unittest.main()
