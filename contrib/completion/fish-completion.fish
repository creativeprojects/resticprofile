#!/usr/bin/env fish
#
# resticprofile fish completion script
# Usage: see https://fishshell.com/docs/current/completions.html#where-to-put-completions

function __resticprofile_completion
    set --local full_cmdline (commandline -x)
    set --local cmdline_to_cursor \
        (string split -- " " (string escape -- (commandline -cx; commandline -ct)))
    set --local current_token_pos (math (count $cmdline_to_cursor) - 1)

    #send commandline to 'resticprofile complete' in the format it expects
    set --local completions ("$full_cmdline[1]" complete "fish:v1" "__POS:$current_token_pos" "$full_cmdline[2..]")

    if test (count $completions) = 0
        return
    end
    
    #handle the directive returned from resticprofile
    switch $completions[-1]
    case "__complete_file"
        set --erase completions[-1]

        set --local file $full_cmdline[-1]
        #if file starts with '-', remove it
        test (string sub --length 1 -- "$file") = "-"; and set file ""

        #do path completion
        set --append completions (__fish_complete_path "$file")

    case "*__complete_restic"
        #string match --regex returns list where first element is the whole string
        #and the rest are capture group matches.
        set --local prefix (string match --regex '^(.*)\.' -- "$completions[-1]")[2]
        set --local suffix (string match --regex '\.([^.]+)$' -- "$completions[-1]")[2]
        set --erase completions[-1]

        #This removes profile prefixes before forwarding the completion to restic
        set --local restic_words
        for word in $cmdline_to_cursor[2..]
            if test (string match --regex ".*\/.*" -- "$word")
                set --append restic_words (string unescape -- "$word")
            else
                #take everything after last dot, if there is one
                set --append restic_words (string match --regex '([^.]+)$' -- $word)[2] 
            end
        end
        
        #Add a space if the cursor is past the last token so that 'complete' doesn't
        #return a command that's already fully typed out
        if test "$restic_words[-1]" = "''"
            set restic_words[-1] " "
        end

        #get restic completions
        set --local restic_values (complete --do-complete "restic $restic_words")

        if test (count $restic_values) = 0
            for x in $completions
                echo $x
            end       
            return
        end

        for value in $restic_values
            #remove empty values
            string match --quiet --regex '^[\r\n\t ]+$' -- "$value"; and continue

            #add prefix back to completion if applicable
            if test -n $prefix && test "$prefix" != "$suffix"
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
