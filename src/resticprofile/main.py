'''
resticprofile main function
'''
from os.path import dirname
from os import environ
from sys import argv, exit
from subprocess import call, DEVNULL
from typing import Union
import toml

from resticprofile import constants
from resticprofile.console import Console
from resticprofile.config import Config
from resticprofile.restic import Restic
from resticprofile.context import Context
from resticprofile.profile import Profile
from resticprofile.filesearch import FileSearch, find_configuration_file, get_default_configuration_locations
from resticprofile.groups import Groups
from resticprofile.error import ConfigError


def main():
    '''
    This is main
    '''
    context = Context(constants.ARGUMENTS_DEFINITION)
    context.load_context_from_command_line(argv)

    console = Console(context.quiet, context.verbose, context.ansi)

    valid_configuration_file = find_configuration_file(context.configuration_file)
    if valid_configuration_file is not None:
        console.debug("Using configuration file " + valid_configuration_file)
        try:
            profiles = toml.load(valid_configuration_file)
        except toml.decoder.TomlDecodeError as err:
            console.error(
                "An error occured while loading the configuration file:")
            console.error(str(err))
            exit(2)
    else:
        console.warning(
            "Configuration file '{}' was not found in either the current directory, home directory or any of these locations:\n{}"
            .format(context.configuration_file, get_default_configuration_locations())
        )
        exit(2)

    base_dir = dirname(valid_configuration_file)
    # todo: allow group of groups
    groups = Groups(profiles)
    if groups.exists(context.profile_name):
        group_name = context.profile_name
        for profile_name in groups.get_profiles(context.profile_name):
            console.debug("Starting profile [{}] from group [{}]".format(profile_name, group_name))
            context.profile_name = profile_name
            run_restic(base_dir, context, profiles, console)
    else:
        # single profile
        run_restic(base_dir, context, profiles, console)


def run_restic(base_dir: str, context: Context, profiles: dict, console: Console):
    file_search = FileSearch(base_dir)
    config = Config(profiles, file_search)
    profile = Profile(config, context.profile_name)
    restic = Restic()
    if context.args:
        # A command was passed as an argument (it has to be the first one after the options)
        restic.command = context.args[0]

    # Build list of arguments to pass to restic
    if constants.SECTION_CONFIGURATION_GLOBAL in profiles:
        context.set_global_context(config)

    try:
        if context.profile_name in profiles:
            profile.set_common_configuration()

            # there's no default command yet
            if not restic.command:
                restic.command = context.default_command

            # we might need the init command so we prepare it
            profile.set_command_configuration('init')

            # we might also need the check command
            profile.set_command_configuration('check')

            # if the command is backup, we need to load the retention model
            if restic.command == constants.COMMAND_BACKUP:
                profile.set_retention_configuration()

            profile.set_command_configuration(restic.command)

            # environment variables
            env = config.get_environment(context.profile_name)
            if env:
                for key, value in env.items():
                    console.debug("Setting environment variable {}".format(key))
                    # When running a group of backups, we might want to put back/remove the environment variable after
                    environ[key] = value

        profile.set_verbosity(context.quiet, context.verbose)
        restic.extend_arguments(profile.get_command_flags(restic.command))

    except FileNotFoundError as error:
        console.error("Error in profile [{}]: {}".format(context.profile_name, str(error)))
        exit(2)

    except ConfigError as error:
        console.error("Error in profile [{}]: {}".format(error.section, error.message))
        exit(2)

    # check that we have the minimum information we need
    if not profile.repository:
        console.error("Error in profile [{}]: a repository is needed in the configuration.".format(context.profile_name))
        exit(2)

    # this is the leftover from the command line
    restic.extend_arguments(context.args[1:])

    if profile.initialize:
        initialize_repository(context, profile, console)

    if profile.run_before:
        run_commands(profile.run_before, console)

    if profile.check_before:
        check(context, profile, console)

    if profile.forget_before:
        cleanup_old_backups(context, profile, console)

    restic_command(context, restic, profile, console)

    if profile.forget_after:
        cleanup_old_backups(context, profile, console)

    if profile.check_after:
        check(context, profile, console)

    if profile.run_after:
        run_commands(profile.run_after, console)


def initialize_repository(context: Context, profile: Profile, console: Console):
    restic_init = Restic(constants.COMMAND_INIT)
    restic_init.extend_arguments(profile.get_command_flags(constants.COMMAND_INIT))
    init_command = context.get_command_prefix() + context.get_restic_path() + " " + restic_init.get_init_command()
    console.info("Initializing repository (if not existing)")
    shell_command(init_command, console, exit_on_returncode=False, display_stderr=False)

def cleanup_old_backups(context: Context, profile: Profile, console: Console):
    restic_retention = Restic(constants.COMMAND_FORGET)
    restic_retention.extend_arguments(profile.get_retention_flags())
    forget_command = context.get_command_prefix() + context.get_restic_path() + " " + restic_retention.get_forget_command()
    console.info("Cleaning up repository using retention information")
    shell_command(forget_command, console)

def check(context: Context, profile: Profile, console: Console):
    restic_check = Restic(constants.COMMAND_CHECK)
    restic_check.extend_arguments(profile.get_command_flags(constants.COMMAND_CHECK))
    check_command = context.get_command_prefix() + context.get_restic_path() + " " + restic_check.get_check_command()
    console.info("Checking repository consistency")
    shell_command(check_command, console)

def restic_command(context: Context, restic: Restic, profile: Profile, console: Console):
    console.info("Starting '{}'".format(restic.command))
    full_command = context.get_command_prefix() + context.get_restic_path() + " " + restic.get_command()
    shell_command(full_command, console, allow_stdin=profile.stdin)
    console.info("Finished '{}'".format(restic.command))

def run_commands(command: Union[str, list], console: Console):
    if isinstance(command, str):
        shell_command(command, console)

    elif isinstance(command, list):
        for single_command in command:
            shell_command(single_command, console)


def shell_command(command: str, console: Console, exit_on_returncode=True, display_stderr=True, allow_stdin=False):
    try:
        stdin_ = DEVNULL
        if allow_stdin:
            stdin_ = None

        stderr_ = None
        if not display_stderr:
            stderr_ = DEVNULL

        console.debug("Starting shell command: " + command)
        returncode = call(command, shell=True, stdin=stdin_, stderr=stderr_)
        if returncode != 0 and exit_on_returncode:
            exit(returncode)

    except KeyboardInterrupt:
        exit()


if __name__ == "__main__":
    main()
