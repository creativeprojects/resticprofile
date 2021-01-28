## Send an email on error (systemd schedule)

In `profiles.yaml` you set:

```yaml
default: 
  ...
  run-after-fail: 
    - 'resticprofile-send-error.sh name@domain.tl'
```

With `/usr/local/bin/resticprofile-send-error.sh` being:

```sh
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
```


See details in [#20](https://github.com/creativeprojects/resticprofile/issues/20)
