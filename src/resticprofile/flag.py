'''
Class Flag
'''
from typing import List, Union

class Flag:
    '''
    Holds flag definition
    '''
    def __init__(self, key: str, value: Union[str, int, bool, list], type_value: str):
        self.key = key
        self.value = value
        self.type = type_value

    def get_flags(self) -> List[str]:
        '''
        Retreive the associated restic flags from the definition
        '''
        flags = []
        if isinstance(self.value, list):
            for value in self.value:
                flag = self.__get_flag_with_value(value)
                if flag:
                    flags.append(flag)

        else:
            # still return a list but with one element
            flag = self.__get_flag_with_value(self.value)
            if flag:
                flags.append(flag)

        return flags

    def __get_flag_with_value(self, value: Union[str, int, bool]) -> str:
        if self.type == 'bool':
            if value:
                return "--{}".format(self.key)
            return ''

        elif self.type in ('str', 'file', 'dir'):
            return "--{} \"{}\"".format(self.key, value.replace('"', '\"'))

        elif self.type == 'int':
            return "--{} {}".format(self.key, value)

        return ''
