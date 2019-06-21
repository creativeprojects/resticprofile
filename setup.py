'''
A setuptools based setup module.
'''
from os import path
import re
from setuptools import setup

NAME = "resticprofile"
SOURCE = 'src'

def read_file(filepath):
    with open(path.join(path.dirname(__file__), filepath)) as fp:
        return fp.read()

def _get_version_match(content):
    # Search for lines of the form: # __version__ = 'ver'
    regex = r"^__version__ = ['\"]([^'\"]*)['\"]"
    version_match = re.search(regex, content, re.M)
    if version_match:
        return version_match.group(1)
    raise RuntimeError("Unable to find version string.")

def get_version(filepath):
    return _get_version_match(read_file(filepath))

setup(
    name=NAME,
    version=get_version(path.join(path.dirname(__file__), SOURCE, NAME, '__init__.py')),
    packages=[NAME],
    package_dir={'': SOURCE},
    include_package_data=True,
    python_requires=">=3.5",
    install_requires=[
        'setuptools',
        'toml>=0.10,<0.11',
        'colorama>=0.4,<0.5'
    ],
    entry_points={
        'console_scripts': [
            'resticprofile=resticprofile.main:main',
        ],
    },
    # metadata to display on PyPI
    author="Fred",
    author_email="Fred@CreativeProjects.Tech",
    description="Manage configuration profiles for restic backup",
    long_description=read_file('README.md'),
    long_description_content_type="text/markdown",
    license="GPL-3.0-or-later",
    keywords="restic backup configuration profiles",
    url="https://github.com/creativeprojects/resticprofile",
    classifiers=[
        "Development Status :: 4 - Beta",
        "Environment :: Console",
        "License :: OSI Approved :: GNU General Public License v3 or later (GPLv3+)",
        "Natural Language :: English",
        "Topic :: System :: Archiving :: Backup",
        "Programming Language :: Python :: 3",
        "Operating System :: OS Independent",
    ]
)
