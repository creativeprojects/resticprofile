#!/usr/bin/env bash
#
# Error notification sendmail script
#
help() {
  cat - <<HELP
  Usage $1 [options] user1@domain user2@domain ...
  Options:
   -s         Only send mail when operating on schedule (RESTICPROFILE_ON_SCHEDULE=1)
   -c command Set the profile command (instead of PROFILE_COMMAND)
   -n name    Set the profile name (instead of PROFILE_NAME)
   -p         Print mail to stdout instead of sending it
   -f         Send mail even when no profile name is specified
HELP
}

# Parse CLI args
FORCE_SENDING=0
SEND_COMMAND=""
while getopts 'c:fhn:ps' flag ; do
  case "${flag}" in
    c) PROFILE_COMMAND="${OPTARG}" ;;
    f) FORCE_SENDING=1 ;;
    n) PROFILE_NAME="${OPTARG}" ;;
    p) SEND_COMMAND="cat -" ;;
    s) (( ${RESTICPROFILE_ON_SCHEDULE:-0} > 0 )) || exit 0 ;;
    *) help "$0" ; exit 0 ;;
  esac
done
shift $((OPTIND-1))

# Parameters
MAIL_TO=""
MAIL_FROM="\"resticprofile $(hostname -f)\" <$USER@$(hostname -f)>"
MAIL_SUBJECT="restic failed: \"${PROFILE_COMMAND}\" in \"${PROFILE_NAME}\""

SEND_COMMAND="${SEND_COMMAND:-sendmail -t}"

DETAILS_COMMAND_RESULT=""
DETAILS_COMMAND=""

# Get command to capture output from scheduler ( if in use )
if [[ -d /etc/systemd/ ]] \
   && (( ${RESTICPROFILE_ON_SCHEDULE:-0} > 0 )) \
   && resticprofile --name "${PROFILE_NAME}" show | grep -v -q -E "scheduler:\s*cron" ; then
  DETAILS_COMMAND="systemctl status --full \"resticprofile-${PROFILE_COMMAND:-*}@profile-${PROFILE_NAME:-*}\""
fi

# Load parameter overrides
RC_FILE="/etc/resticprofile/$(basename "$0").rc}"
[[ -f "${RC_FILE}" ]] && source "${RC_FILE}"

main() {
  if [[ -n "${PROFILE_NAME}" || "${FORCE_SENDING}" == "1" ]] ; then
    if [[ -n "${DETAILS_COMMAND}" ]] ; then
      DETAILS_COMMAND_RESULT="$(${DETAILS_COMMAND})"
    fi

    for email in "$@" "${MAIL_TO}" ; do
      if [[ "${email}" =~ ^[a-zA-Z0-9_.%+-]+@[a-zA-Z0-9_]+[a-zA-Z0-9_.-]+$ ]] ; then
        send_mail "${email}" || echo "Failed sending to \"${email}\""
      elif [[ -n "${email}" ]] ; then
        echo "Skipping notification for invalid address \"${email}\""
      fi
    done
  fi
  return 0
}

send_mail() {
  ${SEND_COMMAND} <<ERRMAIL
To: $1
From: $MAIL_FROM
Subject: $MAIL_SUBJECT
Content-Transfer-Encoding: 8bit
Content-Type: text/plain; charset=UTF-8

${ERROR:-No error information available}

---- 
COMMANDLINE:

${ERROR_COMMANDLINE:-N/A}

----
STDERR:

${ERROR_STDERR:-N/A}

----
DETAILS:

${DETAILS_COMMAND_RESULT:-N/A}

----
CONFIG:

$(resticprofile --name "${PROFILE_NAME}" show)

ERRMAIL
}

# Invoke main and exit without error to ensure other error handlers run
main "$@"
exit 0
