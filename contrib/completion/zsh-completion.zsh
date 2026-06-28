#compdef resticprofile
# resticprofile zsh completion script
#
# Usage:
#   Option 1 - Add to $fpath (recommended, name the file _resticprofile):
#     mkdir -p ~/.zsh/completions
#     resticprofile generate --zsh-completion > ~/.zsh/completions/_resticprofile
#     # Add to ~/.zshrc BEFORE compinit:
#     fpath=(~/.zsh/completions $fpath)
#     autoload -U compinit && compinit
#
#   Option 2 - Source directly (add to ~/.zshrc AFTER compinit):
#     source <(resticprofile generate --zsh-completion)

# _resticprofile_add adds completions, turning an optional tab-separated description
# (sent by "zsh:v2") into a zsh description shown next to each match, like fish does.
# compadd -d is used (rather than _describe) because it adds matches at the same level
# as the plain builtin, so they coexist with the matches added by restic's _restic.
function _resticprofile_add() {
    local -a values display prefixes
    local line value
    integer width=0

    # First pass: find the widest value carrying a description (to align the column)
    for line in "$@"; do
        [[ "${line}" == *$'\t'* ]] || continue
        value="${line%%$'\t'*}"
        (( ${#value} > width )) && width=${#value}
    done

    # Second pass: split out profile prefixes (values ending in ".") and build the
    # display strings for the rest, padding values so descriptions line up.
    for line in "$@"; do
        value="${line%%$'\t'*}"
        if [[ "${value}" == *. ]]; then
            prefixes+=("${value}")
        elif [[ "${line}" == *$'\t'* ]]; then
            values+=("${value}")
            display+=("${(r:width:)value}  -- ${line#*$'\t'}")
        else
            values+=("${value}")
            display+=("${value}")
        fi
    done

    (( ${#values[@]} )) && compadd -d display -- "${values[@]}"
    # Profile prefixes are added with an empty suffix so no space is inserted after
    # the ".", letting "<profile>." be continued with a command (like bash does).
    (( ${#prefixes[@]} )) && compadd -S '' -- "${prefixes[@]}"
}

function _resticprofile() {
    local resticprofile="${words[1]}"

    # Convert zsh's 1-indexed CURRENT to 0-indexed position relative to arguments
    local cursor_pos=$(( CURRENT - 1 ))

    # Get completions from resticprofile ("zsh:v2" enables tab-separated descriptions)
    local -a completions
    completions=("${(@f)$("${resticprofile}" complete "zsh:v2" "__POS:${cursor_pos}" "${words[2,-1]}" 2>/dev/null)}")

    (( ${#completions[@]} == 0 )) && return

    local last="${completions[-1]}"

    if [[ "${last}" == "__complete_file" ]]; then
        completions[-1]=()
        (( ${#completions[@]} )) && _resticprofile_add "${completions[@]}"
        _files
        return
    fi

    if [[ "${last}" == *__complete_restic ]]; then
        # Extract profile prefix (the part before .__complete_restic, if any)
        local profile_prefix=""
        [[ "${last}" != "__complete_restic" ]] && profile_prefix="${last%.__complete_restic}"
        completions[-1]=()

        # Add resticprofile's own completions. These already carry the profile
        # prefix (e.g. "default.show") and must be added before the compset below,
        # while $PREFIX still holds the full "profile." prefixed word.
        (( ${#completions[@]} )) && _resticprofile_add "${completions[@]}"

        # Build args for restic by stripping profile prefixes from the current words
        local -a restic_words=()
        local word
        for word in "${words[2,-1]}"; do
            if [[ "${word}" == */* ]]; then
                restic_words+=("${word}")
            else
                restic_words+=("${word##*.}")
            fi
        done

        # Load restic completion function if not already available
        (( $+functions[_restic] )) || {
            (( $+functions[_completion_loader] )) && _completion_loader restic 2>/dev/null
        }

        if (( $+functions[_restic] )); then
            local saved_service="${service}"
            local -a saved_words=("${words[@]}")
            local saved_current=${CURRENT}

            words=("restic" "${restic_words[@]}")
            CURRENT=${#words[@]}
            service=restic

            # When a profile prefix is present (e.g. completing "default.backup"),
            # strip the "profile." part from the word being completed and move it to
            # $IPREFIX. zsh then matches restic's completions against the bare restic
            # token and automatically re-inserts the "profile." prefix on every match.
            # This works for completions added through _describe / _arguments without
            # having to reimplement compadd's option parsing. ${(b)...} quotes the
            # profile name so it is matched literally even if it contains pattern
            # metacharacters.
            [[ -n "${profile_prefix}" ]] && compset -P "${(b)profile_prefix}."

            _restic

            service="${saved_service}"
            words=("${saved_words[@]}")
            CURRENT=${saved_current}
        fi
        return
    fi

    _resticprofile_add "${completions[@]}"
}

# Register the completion function (works when sourced after compinit)
compdef _resticprofile resticprofile