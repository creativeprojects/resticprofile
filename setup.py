from setuptools import setup, find_packages

setup(
    name="resticprofile",
    version="0.1.1",
    python_requires=">=3.5",
    package_dir={'':'src'},
    packages=find_packages(),
    scripts=['resticprofile.py'],
    # metadata to display on PyPI
    author="Fred",
    author_email="Fred@CreativeProjects.Tech",
    description="Manage configuration profiles for restic backup",
    license="",
    keywords="restic backup configuration profiles",
    url="https://github.com/creativeprojects/resticprofile",
    project_urls={
        "Bug Tracker": "https://github.com/creativeprojects/resticprofile",
        "Documentation": "https://github.com/creativeprojects/resticprofile",
        "Source Code": "https://github.com/creativeprojects/resticprofile",
    }
)
