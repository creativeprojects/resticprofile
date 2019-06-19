from inspect import getsourcefile
from os.path import abspath, isfile, dirname
from os import environ
from sys import argv, exit
from subprocess import call, DEVNULL
import toml

from resticprofile.console import Console
from resticprofile.config import DEFAULTS, ARGUMENTS_DEFINITION, Config
from resticprofile.restic import Restic
from resticprofile.context import Context
from resticprofile.profile import Profile
from resticprofile.filesearch import FileSearch, find_configuration_file, DEFAULT_SEARCH_LOCATIONS


def main():

    context = Context(ARGUMENTS_DEFINITION)
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
            "Configuration file '{}' was not found in either the current directory or these default locations: {}"
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
    if DEFAULTS['global'] in profiles:
        context.set_global_context(config)

    if context.profile_name in profiles:
        profile.set_common_configuration()
        build_argument_list_from_section(restic, context, profiles[context.profile_name])

        # there's no default command yet
        if not restic.command:
            restic.command = context.default_command

        # adds inherited profile (only one at this stage)
        if profile.inherit and restic.command in profiles[profile.inherit]:
            build_argument_list_from_section(restic, context, profiles[profile.inherit][restic.command])

        if restic.command in profiles[context.profile_name]:
            build_argument_list_from_section(restic, context, profiles[context.profile_name][restic.command])

        # inherited environment
        if profile.inherit and DEFAULTS['environment'] in profiles[profile.inherit]:
            env_config = profiles[profile.inherit][DEFAULTS['environment']]
            for key in env_config:
                environ[key.upper()] = env_config[key]
                console.debug("Setting inherited environment variable {}".format(key.upper()))

        if DEFAULTS['environment'] in profiles[context.profile_name]:
            env_config = profiles[context.profile_name][DEFAULTS['environment']]
            for key in env_config:
                environ[key.upper()] = env_config[key]
                console.debug("Setting environment variable {}".format(key.upper()))

    # Clears common arguments and forces them from profile instance
    restic.common_arguments = profile.get_global_flags()

    if context.quiet:
        restic.set_common_argument('--quiet')
    elif context.verbose:
        restic.set_common_argument('--verbose')

    # check that we have the minimum information we need
    if not profile.repository:
        console.error("A repository is needed in the configuration.")
        exit(2)

    restic.extend_arguments(context.args[1:])

    restic_cmd = ""
    for path in ('/usr/bin', '/usr/local/bin', '/opt/local/bin'):
        if isfile(path + '/restic'):
            restic_cmd = path + '/restic'
            break

    command_prefix = ""
    if context.nice:
        command_prefix += context.nice.get_command() + ' '
    if context.ionice:
        command_prefix += context.ionice.get_command() + ' '

    if context.initialize:
        init_command = command_prefix + restic_cmd + " " + restic.get_init_command()
        console.debug(init_command)
        # captures only stdout when we create a new repository; otherwise don't display the error when it exists
        call(init_command, shell=True, stdin=DEVNULL, stderr=DEVNULL)

    if restic.prune_before:
        prune_command = command_prefix + restic_cmd + " " + restic.get_prune_command()
        console.debug(prune_command)
        call(prune_command, shell=True, stdin=DEVNULL)

    full_command = command_prefix + restic_cmd + " " + restic.get_command()
    console.debug(full_command)
    call(full_command, shell=True, stdin=DEVNULL)

    if restic.prune_after:
        prune_command = command_prefix + restic_cmd + " " + restic.get_prune_command()
        console.debug(prune_command)
        call(prune_command, shell=True, stdin=DEVNULL)


def build_argument_list_from_section(restic, context, profiles_section):
    for key in profiles_section:
        if key in ('password-file'):
            # expecting simple string
            if isinstance(profiles_section[key], str):
                value = abspath(profiles_section[key])
                restic.set_common_argument("--{}={}".format(key, value))

        elif key in ('exclude-file', 'tag'):
            # expecting either single string or array of strings
            if isinstance(profiles_section[key], list):
                for value in profiles_section[key]:
                    restic.set_argument("--{}={}".format(key, value))
            elif isinstance(profiles_section[key], str):
                value = profiles_section[key]
                restic.set_argument("--{}={}".format(key, value))

        elif key in ('exclude-caches', 'one-file-system'):
            # expecting boolean value
            if isinstance(profiles_section[key], bool):
                if profiles_section[key]:
                    restic.set_argument("--{}".format(key))

        elif key in ('no-cache'):
            # expecting boolean value
            if isinstance(profiles_section[key], bool):
                if profiles_section[key]:
                    restic.set_common_argument("--{}".format(key))

        elif key == 'repository':
            # expecting single string (and later on, and array of strings!)
            if isinstance(profiles_section[key], str):
                restic.repository = profiles_section[key]
                restic.set_common_argument(
                    "--repo={}".format(profiles_section[key]))

        elif key == 'source':
            # expecting either single string or array of strings
            if isinstance(profiles_section[key], str):
                restic.backup_paths.append(profiles_section[key])
            elif isinstance(profiles_section[key], list):
                for value in profiles_section[key]:
                    restic.backup_paths.append(value)

        elif key in ('initialize'):
            # expecting boolean value
            if isinstance(profiles_section[key], bool):
                context.initialize = profiles_section[key]

        elif key == 'default-command':
            # expecting single string
            if isinstance(profiles_section[key], str):
                context.default_command = profiles_section[key]
                if not restic.command:
                    # also sets the current default command
                    restic.command = profiles_section[key]

        elif key == 'prune-before':
            # expecting boolean value
            if isinstance(profiles_section[key], bool):
                restic.prune_before = profiles_section[key]

        elif key == 'prune-after':
            # expecting boolean value
            if isinstance(profiles_section[key], bool):
                restic.prune_after = profiles_section[key]

        elif key == 'nice':
            context.set_nice(profiles_section)

        elif key == 'ionice':
            context.set_ionice(profiles_section)

        elif key in ('ionice-class', 'ionice-level', 'inherit'):
            # these values are ignored
            pass

        else:
            value = profiles_section[key]
            if isinstance(value, str):
                restic.set_argument("--{}={}".format(key, value))
            elif isinstance(value, int):
                restic.set_argument("--{}={}".format(key, value))
            elif isinstance(value, bool) and value:
                restic.set_argument("--{}".format(key))


if __name__ == "__main__":
    main()
