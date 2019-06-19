import unittest
from resticprofile.nice import Nice

class TestNice(unittest.TestCase):

    def test_can_get_command_with_default_value(self):
        nice = Nice()
        command = nice.get_command()
        self.assertEqual(command, "nice -n 10")

    def test_can_get_command_with_positive_value(self):
        nice = Nice(5)
        command = nice.get_command()
        self.assertEqual(command, "nice -n 5")

    def test_can_get_command_with_negative_value(self):
        nice = Nice(-5)
        command = nice.get_command()
        self.assertEqual(command, "nice -n -5")

    def test_can_get_command_with_value_too_low(self):
        nice = Nice(-100)
        command = nice.get_command()
        self.assertEqual(command, "nice -n -20")

    def test_can_get_command_with_value_too_high(self):
        nice = Nice(100)
        command = nice.get_command()
        self.assertEqual(command, "nice -n 20")

    def test_can_get_command_will_not_throw_exception_and_return_empty_string(self):
        nice = Nice(10, ignore_failure=True)
        command = nice.get_command('Windows')
        self.assertEqual(command, "")

if __name__ == '__main__':
    unittest.main()
