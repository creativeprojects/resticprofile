'''
resticprofile main function
'''
from os.path import isfile, dirname
from os import environ
from sys import argv, exit
from subprocess import call, DEVNULL
import toml

from resticprofile import constants
from resticprofile.console import Console
from resticprofile.config import Config
from resticprofile.restic import Restic
from resticprofile.context import Context
from resticprofile.profile import Profile
from resticprofile.filesearch import FileSearch, find_configuration_file, DEFAULT_SEARCH_LOCATIONS


def main():
    '''
    This is main
    '''
    context = Context(constants.ARGUMENTS_DEFINITION)
    context.load_context_from_command_line(argv)

    console = Console(context.quiet, context.verbose)

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
            .format(context.configuration_file, DEFAULT_SEARCH_LOCATIONS)
        )
        exit(2)

    file_search = FileSearch(dirname(valid_configuration_file))
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

            # if the command is backup, we need to load the retention model
            if restic.command == constants.COMMAND_BACKUP:
                profile.set_retention_configuration()

            profile.set_command_configuration(restic.command)

            # inherited environment
            if profile.inherit:
                if profile.inherit not in profiles:
                    console.error("Error in profile [{}]: inherited profile [{}] was not found.".format(context.profile_name, profile.inherit))
                    exit(2)

                if constants.SECTION_CONFIGURATION_ENVIRONMENT in profiles[profile.inherit]:
                    env_config = profiles[profile.inherit][constants.SECTION_CONFIGURATION_ENVIRONMENT]
                    for key in env_config:
                        environ[key.upper()] = env_config[key]
                        console.debug("Setting inherited environment variable {}".format(key.upper()))

            if constants.SECTION_CONFIGURATION_ENVIRONMENT in profiles[context.profile_name]:
                env_config = profiles[context.profile_name][constants.SECTION_CONFIGURATION_ENVIRONMENT]
                for key in env_config:
                    environ[key.upper()] = env_config[key]
                    console.debug("Setting environment variable {}".format(key.upper()))

        restic.extend_arguments(profile.get_command_flags(restic.command))
    except FileNotFoundError as error:
        console.error("Error in profile [{}]: {}".format(context.profile_name, str(error)))
        exit(2)

    if context.quiet:
        restic.set_common_argument('--quiet')
    elif context.verbose:
        restic.set_common_argument('--verbose')

    # check that we have the minimum information we need
    if not profile.repository:
        console.error("Error in profile [{}]: a repository is needed in the configuration.".format(context.profile_name))
        exit(2)

    restic.extend_arguments(context.args[1:])

    restic_cmd = context.restic_path
    if not restic_cmd:
        for path in ('/usr/bin', '/usr/local/bin', '/opt/local/bin'):
            if isfile(path + '/restic'):
                restic_cmd = path + '/restic'
                break

    command_prefix = ""
    if context.nice:
        command_prefix += context.nice.get_command() + ' '
    if context.ionice:
        command_prefix += context.ionice.get_command() + ' '

    if profile.initialize:
        restic_init = Restic(constants.COMMAND_INIT)
        restic_init.extend_arguments(profile.get_command_flags(constants.COMMAND_INIT))
        init_command = command_prefix + restic_cmd + " " + restic_init.get_init_command()
        console.debug(init_command)
        # captures only stdout when we create a new repository; otherwise don't display the error when it exists
        call(init_command, shell=True, stdin=DEVNULL, stderr=DEVNULL)

    if profile.forget_before:
        restic_retention = Restic(constants.COMMAND_FORGET)
        restic_retention.extend_arguments(profile.get_retention_flags())
        forget_command = command_prefix + restic_cmd + " " + restic_retention.get_forget_command()
        console.debug(forget_command)
        call(forget_command, shell=True, stdin=DEVNULL)

    full_command = command_prefix + restic_cmd + " " + restic.get_command()
    console.debug(full_command)
    call(full_command, shell=True, stdin=DEVNULL)

    if profile.forget_after:
        restic_retention = Restic(constants.COMMAND_FORGET)
        restic_retention.extend_arguments(profile.get_retention_flags())
        forget_command = command_prefix + restic_cmd + " " + restic_retention.get_forget_command()
        console.debug(forget_command)
        call(forget_command, shell=True, stdin=DEVNULL)



if __name__ == "__main__":
    main()
