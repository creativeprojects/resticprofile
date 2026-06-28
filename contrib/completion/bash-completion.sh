#!/usr/bin/env bash
#
# resticprofile bash completion script
# Usage: source this script in your bash profile (or add to bash_completion.d)
function _resticprofile() {
  local resticprofile="${COMP_WORDS[0]}"

  # Pass current cursor position as first argument
  COMP_WORDS[0]="__POS:${COMP_CWORD}"

  # Get completions from resticprofile. Read line by line (mapfile) instead of with
  # word splitting: the restic delegation directive ("bash:v2") may contain tabs that
  # must be preserved.
  local lines=()
  mapfile -t lines < <("$resticprofile" complete "bash:v2" "${COMP_WORDS[@]}")

  # By default every line is a completion candidate
  COMPREPLY=("${lines[@]}")

  # Handle the directive (last line). For restic delegation it also carries the exact
  # arguments to forward to restic, tab-separated:
  #   "[profile.]__complete_restic<TAB>arg1<TAB>arg2..."
  if ((${#lines[@]})) ; then
    local last="${lines[-1]}"
    local directive="${last%%$'\t'*}" # everything before the first tab

    case "${directive}" in
      __complete_file)
        unset 'COMPREPLY[-1]'

        local file="${COMP_WORDS[-1]}"
        [[ "${file:0:1}" == "-" ]] && file=""
        COMPREPLY+=($(compgen -f "$file"))
      ;;

      *__complete_restic)
        # Profile prefix to re-add to restic completions (empty unless completing a
        # "profile.command" token)
        local prefix=""
        [[ "${directive}" != "__complete_restic" ]] && prefix="${directive%.__complete_restic}"

        # Restic arguments to forward, as resolved by resticprofile (own flags removed,
        # any "profile." prefix already stripped). Split on tabs, keeping empty fields
        # (notably the current, not-yet-typed word). A plain "IFS=$'\t' read -a" can't
        # be used: tab is IFS whitespace, so it would drop empty fields.
        local restic_words=()
        if [[ "${last}" == *$'\t'* ]] ; then
          local field
          while IFS= read -r -d $'\t' field || [[ -n "${field}" ]] ; do
            restic_words+=("${field}")
          done < <(printf '%s\t' "${last#*$'\t'}")
        fi

        unset 'COMPREPLY[-1]'

        # Get restic completions
        local restic_values=($(__resticprofile_get_other_completions restic "${restic_words[@]}"))

        if ((${#restic_values[@]})) ; then
          for (( i=0 ; i<${#restic_values[@]} ; i++ )) ; do
            local value="${restic_values[$i]}"

            # Remove any empty values
            if [[ "${value}" =~ (^[\r\n\t ]+$) ]] ; then
              continue
            fi

            # Add profile prefix if requested and append to completions
            if [[ -n $prefix ]] ; then
              COMPREPLY+=("${prefix}.${value}")
            else
              COMPREPLY+=("${value}")
            fi
          done
        fi
      ;;
    esac
  fi

  if [[ ! "${COMPREPLY[*]}" == *. ]]; then
    # Add space if the last character is not a '.'
    compopt +o nospace
  fi
}

# Registering the completion
complete -F _resticprofile -o nospace resticprofile

#
# __resticprofile_get_other_completions
# Author: Brian Beffa <brbsix@gmail.com>
# Original source: https://brbsix.github.io/2015/11/29/accessing-tab-completion-programmatically-in-bash/
# License: LGPLv3 (http://www.gnu.org/licenses/lgpl-3.0.txt)
function __resticprofile_get_other_completions() {
    local completion COMP_CWORD COMP_LINE COMP_POINT COMP_WORDS COMPREPLY=()

    COMP_LINE=$*
    COMP_POINT=${#COMP_LINE}

    eval set -- "$@"

    COMP_WORDS=("$@")

    # add '' to COMP_WORDS if the last character of the command line is a space
    [[ ${COMP_LINE[@]: -1} = ' ' ]] && COMP_WORDS+=('')

    # index of the last word
    COMP_CWORD=$(( ${#COMP_WORDS[@]} - 1 ))

    # determine completion function
    completion=$(complete -p "$1" 2>/dev/null | awk '{print $(NF-1)}')

    # run _completion_loader only if necessary
    [[ -z $completion ]] && declare -F _completion_loader &>/dev/null && {
        # load completion
        _completion_loader "$1"

        # detect completion
        completion=$(complete -p "$1" 2>/dev/null | awk '{print $(NF-1)}')
    }

    # ensure completion was detected
    [[ -n $completion ]] || return 1

    # execute completion function
    "$completion"

    # print completions to stdout
    printf '%s\n' "${COMPREPLY[@]}" | LC_ALL=C sort
}
