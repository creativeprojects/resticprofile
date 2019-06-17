
class Flag:
    def __init__(self, key, value, type_value):
        self.key = key
        self.value = value
        self.type = type_value

    def getFlags(self):
        flags = []
        if isinstance(self.value, list):
            for value in self.value:
                flag = self._getFlagWithValue(value)
                if flag:
                    flags.append(flag)

        else:
            # still return a list but with one element
            flag = self._getFlagWithValue(self.value)
            if flag:
                flags.append(flag)

        return flags

    def _getFlagWithValue(self, value):
        if self.type == 'bool':
            if value:
                return "--{}".format(self.key)
            else:
                return ''
        elif self.type in ('str', 'file'):
            return "--{} '{}'".format(self.key, value)
        elif self.type == 'int':
            return "--{} {}".format(self.key, value)
