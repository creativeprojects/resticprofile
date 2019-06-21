from setuptools import setup, find_packages

setup(
    name="resticprofile",
    version="0.1.1",
    python_requires=">=3.5",
    package_dir={'':'src'},
    packages=find_packages(exclude=["tests"]),
    # metadata to display on PyPI
    author="Fred",
    author_email="Fred@CreativeProjects.Tech",
    description="Manage configuration profiles for restic backup",
    license="GPL-3.0-or-later",
    keywords="restic backup configuration profiles",
    url="https://github.com/creativeprojects/resticprofile",
    project_urls={
        "Bug Tracker": "https://github.com/creativeprojects/resticprofile",
        "Documentation": "https://github.com/creativeprojects/resticprofile",
        "Source Code": "https://github.com/creativeprojects/resticprofile",
    }
)
