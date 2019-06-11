from inspect import getsourcefile
from os.path import abspath, isfile, dirname
from os import chdir, getcwd, environ
from getopt import getopt, GetoptError
from sys import argv, exit
from subprocess import call
from ast import literal_eval
from lib.console import Console
import configparser

defaults = {
    'config_file': 'profiles.conf',
    'profile': 'default',
    'global': 'global',
    'separator': '.',
    'environment': 'env',
}

arguments_definition = {
    'help': {
        'short': 'h',
        'long': 'help',
        'argument': False,
    },
    'quiet': {
        'short': 'q',
        'long': 'quiet',
        'argument': False,
    },
    'verbose': {
        'short': 'v',
        'long': 'verbose',
        'argument': False,
    },
    'config': {
        'short': 'c',
        'long': 'config',
        'argument': True,
        'argument_name': 'configuration_file'
    },
    'name': {
        'short': 'n',
        'long': 'name',
        'argument': True,
        'argument_name': 'profile_name'
    }
}


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
    verbose = None
    quiet = None

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

    config = configparser.ConfigParser()
    if valid_configuration_file != None:
        print("Using configuration file " + valid_configuration_file)
        config.read(valid_configuration_file)
    else:
        print("Configuration file '" + configuration_file + "' was not found in '" + current_directory + "' and '" + script_directory + "'")
        config[configuration_name] = {}

    restic_arguments = []
    backup_paths = []
    # Build list of arguments to pass to restic
    if configuration_name in config:
        build_argument_list_from_section(restic_arguments, backup_paths, config[configuration_name])

    if configuration_name + defaults['separator'] + restic_command in config:
        build_argument_list_from_section(restic_arguments, backup_paths, config[configuration_name + defaults['separator'] + restic_command])

    if configuration_name + defaults['separator'] + defaults['environment'] in config:
        env_config = config[configuration_name + defaults['separator'] + defaults['environment']]
        for key in env_config:
            environ[key.upper()] = env_config[key]
            print("Setting environment variable {}".format(key.upper()))

    if quiet:
        restic_arguments.append('--quiet')
    elif verbose:
        restic_arguments.append('--verbose')

    restic_arguments.extend(args[1:])

    restic_cmd = ""
    for path in ('/usr/bin', '/usr/local/bin', '/opt/local/bin'):
        if isfile(path + '/restic'):
            restic_cmd = path + '/restic'
            break

    full_command = restic_cmd + " " + restic_command + " " + ' '.join(restic_arguments) + " " + ' '.join(backup_paths)
    console.debug(full_command)
    call(full_command, shell=True)

def build_argument_list_from_section(restic_arguments, backup_paths, config_section):
    for key in config_section:
        if key in ('exclude-file', 'password-file'):
            value = abspath(config_section[key])
            restic_arguments.append("--" + key + '=' + value)
        elif key in ('tag'):
            values = literal_eval(config_section[key])
            for value in values:
                restic_arguments.append("--" + key + '=' + value)
        elif key in ('exclude-caches', 'one-file-system', 'no-cache'):
            value = config_section.getboolean(key)
            if value:
                restic_arguments.append("--" + key)
        elif key == 'repository':
            restic_arguments.append("-r=" + config_section[key])
        elif key == 'backup':
            values = literal_eval(config_section[key])
            for value in values:
                backup_paths.append(value)
        else:
            value = config_section[key]
            restic_arguments.append("--" + key + '=' + value)

if __name__ == "__main__":
    main()
