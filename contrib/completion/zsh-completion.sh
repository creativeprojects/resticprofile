#!/usr/bin/env zsh
#
# resticprofile zsh completion script
# Usage: source this script in your zsh profile (or add to zsh_completion.d)

# TODO: This script redirects to the bash_completion script
# TODO: A dedicated completion for zsh should be created instead

autoload -U +X compinit && compinit
autoload -U +X bashcompinit && bashcompinit
eval "$(resticprofile completion-script --bash)"

#
# Disable other completions for now
function __resticprofile_get_other_completions() {
  return
}
