from inspect import getsourcefile
from os.path import abspath, isfile, dirname
from os import chdir, getcwd, environ
from getopt import getopt, GetoptError
from sys import argv, exit
from subprocess import call
from ast import literal_eval
from lib.console import Console
from lib.config import defaults, arguments_definition
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
    if len(args) > 0:
        restic_command = args[0]

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

    repository = ""
    restic_arguments = []
    backup_paths = []
    # Build list of arguments to pass to restic
    if defaults['global'] in profiles:
        build_argument_list_from_section(repository, restic_arguments, backup_paths, global_config, profiles[defaults['global']])

    if configuration_name in profiles:
        build_argument_list_from_section(repository, restic_arguments, backup_paths, global_config, profiles[configuration_name])

    if configuration_name + defaults['separator'] + restic_command in profiles:
        build_argument_list_from_section(repository, restic_arguments, backup_paths, global_config, profiles[configuration_name + defaults['separator'] + restic_command])

    if configuration_name + defaults['separator'] + defaults['environment'] in profiles:
        env_config = profiles[configuration_name + defaults['separator'] + defaults['environment']]
        for key in env_config:
            environ[key.upper()] = env_config[key]
            print("Setting environment variable {}".format(key.upper()))

    if quiet:
        restic_arguments.append('--quiet')
    elif verbose:
        restic_arguments.append('--verbose')

    # check that we have the minimum information we need
    if not repository:
        console.error("A repository is needed in the configuration.")
        exit(2)

    restic_arguments.extend(args[1:])

    restic_cmd = ""
    for path in ('/usr/bin', '/usr/local/bin', '/opt/local/bin'):
        if isfile(path + '/restic'):
            restic_cmd = path + '/restic'
            break

    full_command = restic_cmd + " " + restic_command + " " + ' '.join(restic_arguments) + " " + ' '.join(backup_paths)
    console.debug(full_command)
    call(full_command, shell=True)

def build_argument_list_from_section(repository, restic_arguments, backup_paths, global_config, profiles_section):
    for key in profiles_section:
        if key in ('password-file'):
            # expecting simple string
            if profiles_section[key] is str:
                value = abspath(profiles_section[key])
                restic_arguments.append("--" + key + '=' + value)

        elif key in ('exclude-file', 'tag'):
            # expecting either single string or array of strings
            if profiles_section[key] is list:
                for value in profiles_section[key]:
                    restic_arguments.append("--" + key + '=' + value)
            elif profiles_section[key] is str:
                restic_arguments.append("--" + key + '=' + value)

        elif key in ('exclude-caches', 'one-file-system', 'no-cache'):
            # expecting boolean value
            if profiles_section[key] is bool:
                if profiles_section[key]:
                    restic_arguments.append("--" + key)

        elif key == 'repository':
            # expecting single string (and later on, and array of strings!)
            if profiles_section[key] is str:
                repository = profiles_section[key]
                restic_arguments.append("-r=" + profiles_section[key])

        elif key == 'backup':
            # expecting either single string or array of strings
            if profiles_section[key] is str:
                backup_paths.append(profiles_section[key])
            elif profiles_section[key] is list:
                for value in profiles_section[key]:
                    backup_paths.append(value)

        elif key in ('initialize', 'ionice'):
            # expecting boolean value
            if profiles_section[key] is bool:
                global_config[key] = profiles_section[key]

        elif key == 'default-command':
            # expecting single string
            if profiles_section[key] is str:
                global_config[key] = profiles_section[key]

        else:
            value = profiles_section[key]
            if value is str:
                restic_arguments.append("--" + key + '=' + value)
            elif value is int:
                restic_arguments.append("--" + key + '=' + value)

if __name__ == "__main__":
    main()
