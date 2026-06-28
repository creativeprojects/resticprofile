#!/usr/bin/env fish
#
# resticprofile fish completion script
# Usage: see https://fishshell.com/docs/current/completions.html#where-to-put-completions

function __resticprofile_completion
    set --local full_cmdline
    set --local cmdline_to_cursor_unsplit

    #handle deprecated option in fish v4
    if test (string split -- "." "$FISH_VERSION")[1] -gt 3
        set full_cmdline (commandline -x)
        set cmdline_to_cursor_unsplit (commandline -cx; commandline -ct)
    else 
        set full_cmdline (commandline -o)
        set cmdline_to_cursor_unsplit (commandline -co; commandline -ct)
    end

    set --local cmdline_to_cursor \
        (string split -- " " (string escape -- $cmdline_to_cursor_unsplit))

    set --local current_token_pos (math (count $cmdline_to_cursor) - 1)

    #send commandline to 'resticprofile complete' in the format it expects
    #("fish:v2" makes resticprofile resolve the restic arguments to forward).
    #$full_cmdline[2..] is unquoted so each token is passed as a separate argument;
    #quoting would join them into one (e.g. "-n default ba"), breaking the parsing.
    set --local completions ("$full_cmdline[1]" complete "fish:v2" "__POS:$current_token_pos" $full_cmdline[2..])

    if test (count $completions) = 0
        return
    end

    #The last line is the directive. For restic delegation (fish:v2) it also carries
    #the exact arguments to forward to restic, tab-separated:
    #  "[profile.]__complete_restic<TAB>arg1<TAB>arg2..."
    set --local directive_parts (string split \t -- $completions[-1])
    set --local directive $directive_parts[1]

    #handle the directive returned from resticprofile
    switch $directive
    case "__complete_file"
        set --erase completions[-1]

        set --local file $full_cmdline[-1]
        #if file starts with '-', remove it
        test (string sub --length 1 -- "$file") = "-"; and set file ""

        #do path completion
        set --append completions (__fish_complete_path "$file")

    case "*__complete_restic"
        #profile prefix to re-add to restic completions (empty unless completing a
        #"profile.command" token)
        set --local prefix (string match --regex '^(.*)\.__complete_restic$' -- "$directive")[2]
        #restic arguments to forward, as resolved by resticprofile (own flags removed,
        #any "profile." prefix already stripped from the command)
        set --local restic_words $directive_parts[2..]
        set --erase completions[-1]

        #build a "restic ..." command line and ask restic for its completions. An empty
        #word (the current, not-yet-typed token) becomes a trailing space so restic
        #proposes the next token; other words are escaped to survive re-parsing.
        set --local restic_cmd restic
        for word in $restic_words
            if test -z "$word"
                set restic_cmd "$restic_cmd "
            else
                set restic_cmd "$restic_cmd "(string escape -- "$word")
            end
        end
        set --local restic_values (complete --do-complete "$restic_cmd")

        for value in $restic_values
            #remove empty values
            string match --quiet --regex '^[\r\n\t ]+$' -- "$value"; and continue

            #add prefix back to completion if applicable
            if test -n "$prefix"
                set --append completions "$prefix.$value"
            else
                set --append completions "$value"
            end
        end
    end

    for x in $completions
        echo $x
    end

end

complete --command resticprofile --no-files --arguments "(__resticprofile_completion)"

