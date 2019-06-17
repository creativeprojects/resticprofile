from inspect import getsourcefile
from os.path import abspath, isfile, dirname
from os import chdir, getcwd, environ
from getopt import getopt, GetoptError
from sys import argv, exit
from subprocess import call, DEVNULL
from ast import literal_eval
from lib.console import Console
from lib.config import defaults, arguments_definition
from lib.restic import Restic
from lib.context import Context
from lib.nice import Nice
from lib.ionice import IONice
from lib.profile import Profile
import toml

def get_short_options():
    short_options = ""
    for name, options in arguments_definition.items():
        short_options += options['short'] + (":" if options['argument'] else "")
    return short_options

def get_long_options():
    long_options = []
    for name, options in arguments_definition.items():
        long_options.append(options['long'] + ("=" if options['argument'] else ""))
    return long_options

def get_possible_options_for(option):
    return [ "-" + arguments_definition[option]['short'], "--" + arguments_definition[option]['long'] ]

def main():
    try:
        short_options = get_short_options()
        long_options = get_long_options()
        opts, args = getopt(argv[1:], short_options, long_options)

    except GetoptError as err:
        console = Console()
        console.error("Error in the command arguments: " + err.msg)
        console.usage(argv[0])
        exit(2)

    context = Context()

    for option, argument in opts:
        if option in get_possible_options_for('help'):
            Console().usage(argv[0])
            exit()
        elif option in get_possible_options_for('quiet'):
            context.quiet = True

        elif option in get_possible_options_for('verbose'):
            context.verbose = True

        elif option in get_possible_options_for('config'):
            context.configuration_file = argument

        elif option in get_possible_options_for('name'):
            context.profile_name = argument

        else:
            assert False, "unhandled option"
    
    console = Console(context.quiet, context.verbose)

    # Current directory where the script was started from
    current_directory = getcwd()
    # Directory where the script is located
    script_directory = dirname(abspath(getsourcefile(lambda:0)))     # Who said python reads like English?

    valid_configuration_file = None
    for file in (current_directory + "/" + context.configuration_file, script_directory + "/" + context.configuration_file):
        if isfile(file):
            valid_configuration_file = file
            chdir(dirname(valid_configuration_file))
            break


    if valid_configuration_file != None:
        console.debug("Using configuration file " + valid_configuration_file)
        try:
            profiles = toml.load(valid_configuration_file)
        except toml.decoder.TomlDecodeError as err:
            console.error("An error occured while loading the configuration file:")
            console.error(str(err))
            exit(2)
    else:
        console.warning("Configuration file '" + context.configuration_file + "' was not found in either current or script directory.")
        exit(2)

    backup = Profile(context.profile_name)
    restic = Restic()
    if len(args) > 0:
        # A command was passed as an argument (it has to be the first one after the options)
        restic.command = args[0]

    # Build list of arguments to pass to restic
    if defaults['global'] in profiles:
        context.set_global_context(profiles[defaults['global']])

    if context.profile_name in profiles:
        backup.set_global_configuration(profiles[context.profile_name])
        build_argument_list_from_section(restic, context, profiles[context.profile_name])

        # there's no default command yet
        if not restic.command:
            restic.command = context.default_command

        if restic.command in profiles[context.profile_name]:
            build_argument_list_from_section(restic, context, profiles[context.profile_name][restic.command])

        if defaults['environment'] in profiles[context.profile_name]:
            env_config = profiles[context.profile_name][defaults['environment']]
            for key in env_config:
                environ[key.upper()] = env_config[key]
                console.debug("Setting environment variable {}".format(key.upper()))

    # Clears common arguments and forces them from backup instance
    restic.common_arguments = backup.get_global_flags()

    if context.quiet:
        restic.set_common_argument('--quiet')
    elif context.verbose:
        restic.set_common_argument('--verbose')

    # check that we have the minimum information we need
    if not restic.repository:
        console.error("A repository is needed in the configuration.")
        exit(2)

    restic.extend_arguments(args[1:])

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
        call(init_command, shell = True, stdin = DEVNULL, stderr = DEVNULL)

    if restic.prune_before:
        prune_command = command_prefix + restic_cmd + " " + restic.get_prune_command()
        console.debug(prune_command)
        call(prune_command, shell = True, stdin = DEVNULL)

    full_command = command_prefix + restic_cmd + " " + restic.get_command()
    console.debug(full_command)
    call(full_command, shell = True, stdin = DEVNULL)

    if restic.prune_after:
        prune_command = command_prefix + restic_cmd + " " + restic.get_prune_command()
        console.debug(prune_command)
        call(prune_command, shell = True, stdin = DEVNULL)

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
                restic.set_common_argument("--repo={}".format(profiles_section[key]))

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

        elif key in ('ionice-class', 'ionice-level'):
            # these values are ignored
            None

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
