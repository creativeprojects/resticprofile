#!/usr/bin/env bash
#
# resticprofile bash completion script
# Usage: source this script in your bash profile (or add to bash_completion.d)
function _resticprofile() {
  local resticprofile="${COMP_WORDS[0]}"

  # Pass current cursor position as first argument
  COMP_WORDS[0]="__POS:${COMP_CWORD}"

  # Get completions from resticprofile
  COMPREPLY=($("$resticprofile" complete "bash:v1" "${COMP_WORDS[@]}"))

  # Handle completion requests (last item in result of prev command)
  if ((${#COMPREPLY[@]})) ; then
    case "${COMPREPLY[-1]}" in
      __complete_file)
        unset COMPREPLY[-1]

        local file="${COMP_WORDS[-1]}"
        [[ "${file:0:1}" == "-" ]] && file=""
        COMPREPLY+=($(compgen -f "$file"))
      ;;

      **__complete_restic)
        local prefix="${COMPREPLY[-1]%.*}"  # everything before last dot
        local suffix="${COMPREPLY[-1]##*.}" # everything after last dot
        unset COMPREPLY[-1]

        # Remove profile prefixes before passing args to restic (removes any [prefix.]value, keeps paths)
        local restic_words=()
        for (( i=1 ; i<${#COMP_WORDS[@]} ; i++ )) ; do
          local word="${COMP_WORDS[$i]}"
          if [[ "${word}" =~ (.*/.*) ]] ; then
            restic_words+=("${word}")
          else
            restic_words+=("${word##*.}")
          fi
        done

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
            if [[ -n $prefix && "$prefix" != "$suffix" ]] ; then
              COMPREPLY+=("${prefix}.${value}")
            else
              COMPREPLY+=("${value}")
            fi
          done
        fi
      ;;
    esac
  fi
}

# Registering the completion
complete -F _resticprofile resticprofile

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
