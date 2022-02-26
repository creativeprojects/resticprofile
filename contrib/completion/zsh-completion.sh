#!/usr/bin/env zsh
#
# resticprofile zsh completion script
# Usage: source this script in your zsh profile (or add to zsh_completion.d)

# TODO: This script redirects to the bash_completion script
# TODO: A dedicated completion for zsh should be created instead

autoload -U +X compinit && compinit
autoload -U +X bashcompinit && bashcompinit

# Continue only if bashcompinit succeeded
declare -F complete &>/dev/null || exit 1

# Load bash script
__resticprofile_script=""

if [[ -x "$(which resticprofile)" ]] ; then
  __resticprofile_script="$(resticprofile completion-script --bash)"
else
  echo "resticprofile not found or not executable"
  exit 1
fi

if echo "${__resticprofile_script}" | grep -q "__resticprofile_get_other_completions()" ; then
  eval "${__resticprofile_script}"
else
  echo "resticprofile did not return a completion script"
  exit 1
fi

#
# Disable other completions for now
function __resticprofile_get_other_completions() {
  return
}
