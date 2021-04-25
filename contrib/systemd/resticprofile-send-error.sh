#!/usr/bin/env bash
[[ -z "${PROFILE_NAME}" ]] || sendmail -t <<ERRMAIL
To: $1
From: "resticprofile $(hostname -f)" <$USER@$(hostname -f)>
Subject: restic failed: ${PROFILE_COMMAND} "${PROFILE_NAME}"
Content-Transfer-Encoding: 8bit
Content-Type: text/plain; charset=UTF-8

${ERROR}

---- 
COMMANDLINE:

${ERROR_COMMANDLINE}

----
STDERR:

${ERROR_STDERR}

----
DETAILS:

$(systemctl status --full "resticprofile-${PROFILE_COMMAND}@profile-${PROFILE_NAME}")

----
CONFIG:

$(resticprofile --name "${PROFILE_NAME}" show)

ERRMAIL
exit 0
