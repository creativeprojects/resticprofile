#!/usr/bin/env bash
[[ -z "${PROFILE_NAME}" ]] || sendmail -t <<ERRMAIL
To: $1
From: "Resticprofile $(hostname -f)" <$USER@$(hostname -f)>
Subject: Restic Failed: ${PROFILE_COMMAND} "${PROFILE_NAME}"
Content-Transfer-Encoding: 8bit
Content-Type: text/plain; charset=UTF-8

${ERROR}

---- 
DETAILS:

$(systemctl status --full "resticprofile-${PROFILE_COMMAND}@profile-${PROFILE_NAME}")
ERRMAIL
exit 0
