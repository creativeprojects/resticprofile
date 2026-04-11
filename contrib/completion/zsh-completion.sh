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

function _resticprofile() {
    local resticprofile="${words[1]}"

    # Convert zsh's 1-indexed CURRENT to 0-indexed position relative to arguments
    local cursor_pos=$(( CURRENT - 1 ))

    # Get completions from resticprofile
    local -a completions
    completions=("${(@f)$("${resticprofile}" complete "zsh:v1" "__POS:${cursor_pos}" "${words[2,-1]}" 2>/dev/null)}")

    (( ${#completions[@]} == 0 )) && return

    local last="${completions[-1]}"

    if [[ "${last}" == "__complete_file" ]]; then
        completions[-1]=()
        (( ${#completions[@]} )) && compadd -- "${completions[@]}"
        _files
        return
    fi

    if [[ "${last}" == *__complete_restic ]]; then
        # Extract profile prefix (the part before .__complete_restic, if any)
        local profile_prefix=""
        [[ "${last}" != "__complete_restic" ]] && profile_prefix="${last%.__complete_restic}"
        completions[-1]=()

        # Add any other completions already collected
        (( ${#completions[@]} )) && compadd -- "${completions[@]}"

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

            if [[ -n "${profile_prefix}" ]]; then
                # Override compadd to prepend the profile prefix to each completion.
                # Note: this shadows the compadd builtin for the duration of _restic's execution.
                local _rp_pfx="${profile_prefix}."

                function compadd() {
                    local -a _opts=() _items=()
                    local -i _after_dash=0
                    local _next_opt=""

                    for _arg in "$@"; do
                        if [[ -n "${_next_opt}" ]]; then
                            if [[ "${_next_opt}" == "-a" ]]; then
                                # Items come from a named array; expand it inline
                                eval "_items+=(\"\${${_arg}[@]}\")"
                                _opts[-1]=()  # Remove the -a option we already added
                            else
                                _opts+=("${_arg}")
                            fi
                            _next_opt=""
                        elif (( _after_dash )); then
                            _items+=("${_arg}")
                        elif [[ "${_arg}" == "--" ]]; then
                            _after_dash=1
                        elif [[ "${_arg}" == "-a" ]]; then
                            _opts+=("${_arg}")
                            _next_opt="${_arg}"
                        elif [[ "${_arg}" == -[PpSsdGWFXJVEo] ]]; then
                            # Options that take a following value
                            _opts+=("${_arg}")
                            _next_opt="${_arg}"
                        elif [[ "${_arg}" == -* || "${_arg}" == +* ]]; then
                            _opts+=("${_arg}")
                        else
                            _items+=("${_arg}")
                        fi
                    done

                    local -a _prefixed=()
                    for _item in "${_items[@]}"; do
                        _prefixed+=("${_rp_pfx}${_item}")
                    done
                    builtin compadd "${_opts[@]}" -- "${_prefixed[@]}"
                }

                { _restic; } always { unfunction compadd 2>/dev/null }
            else
                _restic
            fi

            service="${saved_service}"
            words=("${saved_words[@]}")
            CURRENT=${saved_current}
        fi
        return
    fi

    compadd -- "${completions[@]}"
}

# Register the completion function (works when sourced after compinit)
compdef _resticprofile resticprofile