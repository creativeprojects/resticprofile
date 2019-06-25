'''
resticprofile help
'''

def get_options_help(args_definition):
    args_help = []
    for name, options in args_definition.items():
        option_help = ""
        args = []
        if 'short' in options:
            args.append('-' + options['short'])
        if 'long' in options:
            argument = ""
            if 'argument' in options and 'argument_name' in options and options['argument']:
                argument = " <" + options['argument_name'] + ">"
            args.append('--' + options['long'] + argument)
        if args:
            option_help = "[{}]".format('|'.join(args))
        if option_help:
            args_help.append(option_help)
    # sort the list because it's always built in random order for some reason
    args_help.sort()
    return args_help
