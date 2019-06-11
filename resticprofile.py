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
import toml

def get_options_help():
    for name, options in arguments_definition.items():
        option_help = "[-" + options['short']
        option_help += "|--" + options['long']
        option_help += " <" + options['argument_name'] + ">" if options['argument'] else ""
        option_help += "]"
        yield option_help

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

def usage():
    print("Usage:")
    print(" " + argv[0] + " " + "\n   ".join(get_options_help()) + " COMMAND")
    print
    print("Default configuration file is: '{}' (in the current folder)".format(defaults['config_file']))
    print("Default configuration profile is: {}".format(defaults['profile']))
    print

def main():
    try:
        short_options = get_short_options()
        long_options = get_long_options()
        opts, args = getopt(argv[1:], short_options, long_options)

    except GetoptError as err:
        Console().error("Error in the command arguments: " + err.msg) # will print something like "option -a not recognized"
        usage()
        exit(2)

    configuration_file = defaults['config_file']
    configuration_name = defaults['profile']
    verbose = defaults['verbose']
    quiet = defaults['quiet']
    global_config = {}
    global_config['ionice'] = defaults['ionice']
    global_config['default-command'] = defaults['default-command']
    global_config['initialize'] = defaults['initialize']

    for option, argument in opts:
        if option in get_possible_options_for('help'):
            usage()
            exit()
        elif option in get_possible_options_for('quiet'):
            quiet = True

        elif option in get_possible_options_for('verbose'):
            verbose = True

        elif option in get_possible_options_for('config'):
            configuration_file = argument

        elif option in get_possible_options_for('name'):
            configuration_name = argument

        else:
            assert False, "unhandled option"
    
    console = Console(quiet, verbose)

    restic_command = "snapshots"

    # Current directory where the script was started from
    current_directory = getcwd()
    # Directory where the script is located
    script_directory = dirname(abspath(getsourcefile(lambda:0)))     # Who said python reads like English?

    valid_configuration_file = None
    for file in (current_directory + "/" + configuration_file, script_directory + "/" + configuration_file):
        if isfile(file):
            valid_configuration_file = file
            chdir(dirname(valid_configuration_file))
            break


    if valid_configuration_file != None:
        console.debug("Using configuration file " + valid_configuration_file)
        profiles = toml.load(valid_configuration_file)
    else:
        console.warning("Configuration file '" + configuration_file + "' was not found in either current or script directory.")
        exit(2)

    restic = Restic()
    if len(args) > 0:
        # A command was passed as an argument (it has to be the first one after the options)
        restic.command = args[0]

    # Build list of arguments to pass to restic
    if defaults['global'] in profiles:
        build_argument_list_from_section(restic, global_config, profiles[defaults['global']])

    if configuration_name in profiles:
        build_argument_list_from_section(restic, global_config, profiles[configuration_name])

        # there's no default command yet
        if not restic.command:
            restic.command = defaults['default-command']

        if restic.command in profiles[configuration_name]:
            build_argument_list_from_section(restic, global_config, profiles[configuration_name][restic.command])

        if defaults['environment'] in profiles[configuration_name]:
            env_config = profiles[configuration_name][defaults['environment']]
            for key in env_config:
                environ[key.upper()] = env_config[key]
                console.debug("Setting environment variable {}".format(key.upper()))

    if quiet:
        restic.set_common_argument('--quiet')
    elif verbose:
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

    if global_config['initialize']:
        init_command = restic_cmd + " " + restic.get_init_command()
        console.debug(init_command)
        # captures only stdout when we create a new repository; otherwise don't display the error when it exists
        call(init_command, shell = True, stdin = DEVNULL, stderr = DEVNULL)

    if restic.prune_before:
        prune_command = restic_cmd + " " + restic.get_prune_command()
        console.debug(prune_command)
        call(prune_command, shell = True, stdin = DEVNULL)

    full_command = restic_cmd + " " + restic.get_command()
    console.debug(full_command)
    call(full_command, shell = True, stdin = DEVNULL)

    if restic.prune_after:
        prune_command = restic_cmd + " " + restic.get_prune_command()
        console.debug(prune_command)
        call(prune_command, shell = True, stdin = DEVNULL)

def build_argument_list_from_section(restic, global_config, profiles_section):
    for key in profiles_section:
        if key in ('password-file'):
            # expecting simple string
            if isinstance(profiles_section[key], str):
                value = abspath(profiles_section[key])
                restic.set_common_argument("--" + key + '=' + value)

        elif key in ('exclude-file', 'tag'):
            # expecting either single string or array of strings
            if isinstance(profiles_section[key], list):
                for value in profiles_section[key]:
                    restic.set_argument("--" + key + '=' + value)
            elif isinstance(profiles_section[key], str):
                restic.set_argument("--" + key + '=' + value)

        elif key in ('exclude-caches', 'one-file-system', 'no-cache'):
            # expecting boolean value
            if isinstance(profiles_section[key], bool):
                if profiles_section[key]:
                    restic.set_argument("--" + key)

        elif key == 'repository':
            # expecting single string (and later on, and array of strings!)
            if isinstance(profiles_section[key], str):
                restic.repository = profiles_section[key]
                restic.set_common_argument("-r=" + profiles_section[key])

        elif key == 'backup':
            # expecting either single string or array of strings
            if isinstance(profiles_section[key], str):
                restic.backup_paths.append(profiles_section[key])
            elif isinstance(profiles_section[key], list):
                for value in profiles_section[key]:
                    restic.backup_paths.append(value)

        elif key in ('initialize', 'ionice'):
            # expecting boolean value
            if isinstance(profiles_section[key], bool):
                global_config[key] = profiles_section[key]

        elif key == 'default-command':
            # expecting single string
            if isinstance(profiles_section[key], str):
                global_config[key] = profiles_section[key]
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

        else:
            value = profiles_section[key]
            if isinstance(value, str):
                restic.set_argument("--" + key + '=' + value)
            elif isinstance(value, int):
                restic.set_argument("--" + key + '=' + value)

if __name__ == "__main__":
    main()
