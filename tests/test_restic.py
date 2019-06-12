import unittest
from src.lib.restic import Restic

class TestRestic(unittest.TestCase):

    def test_can_get_command_with_no_argument(self):
        restic = Restic()
        restic.command = "backup"
        command = restic.get_command()
        self.assertEqual(command, "backup")

if __name__ == '__main__':
    unittest.main()
