#!/usr/bin/env sh
ROOT_PATH="${ROOT_PATH:-}"

FILES_OWNER="root:root"
if [ -n "${ROOT_PATH}" ] && [ "${ROOT_PATH}" != "/usr/local" ] ; then
  FILES_OWNER="$(id -u):$(id -g)"
fi

# Temp dir (using fixed path to ensure "rm -rf" will not have side effects)
TEMP_PATH="${ROOT_PATH}/tmp/.resticprofile-setup"
if [ -d "$TEMP_PATH" ] ; then
  rm -rf "${TEMP_PATH}"
fi
mkdir -p "$TEMP_PATH"
trap "rm -rf \"${TEMP_PATH}\"" EXIT INT TERM

# Paths
CACHE_PATH="${TEMP_PATH}/cache"
CONFIG_PATH="${ROOT_PATH}/etc/resticprofile"
CONFIG_CACHE_FILE="${CONFIG_PATH}/.dist.cache"
REPOSITORY_SECRET="repositories.d/default.secret"
SECRET_FILE="${CONFIG_PATH}/${REPOSITORY_SECRET}"

# Config files that are merged with .dist files when already existing
MERGEABLES="conf.d/backup.conf conf.d/check.conf conf.d/hooks.conf"
MERGEABLES="${MERGEABLES} conf.d/metrics.conf conf.d/prune.conf"
MERGEABLES="${MERGEABLES} profiles.conf repositories.d/default.conf"

# Search path of dirs to install shell completions (only first match will be used)
COMPLETION_DIRS="${ROOT_PATH}/usr/share/bash-completion/completions"
COMPLETION_DIRS="${COMPLETION_DIRS} ${ROOT_PATH}/usr/share/bash-completion/bash_completion"
COMPLETION_DIRS="${COMPLETION_DIRS} ${ROOT_PATH}/usr/local/etc/bash_completion.d"
COMPLETION_DIRS="${COMPLETION_DIRS} ${ROOT_PATH}/etc/bash_completion.d"

# Fix permissions (only root may edit and read since password & tokens can be in any of the files)
set_permission() {
  _path="${CONFIG_PATH}/$1"

  if [ -e "${_path}" ] ; then
    echo "Setting perms on ${_path}"
    chown "${FILES_OWNER}" "${_path}" || return 1
  fi

  if [ -d "${_path}" ] || echo "${_path}" | grep -q -E '.rc$' ; then
    chmod 0755 "${_path}"
  elif [ -f "${_path}" ] ; then
    echo "$1" | grep -q .secret \
      && chmod 0400 "${_path}" \
      || chmod 0640 "${_path}"
  fi
}

if cd "${CONFIG_PATH}" ; then
  for file in *.dist \
              conf.d conf.d/*.dist \
              profiles.d profiles.d/*.dist \
              repositories.d repositories.d/*.dist ${REPOSITORY_SECRET} \
              templates templates/*.dist \
              ${MERGEABLES} ; do
    set_permission "${file}"
  done
else
  echo "config path (${CONFIG_PATH}) not found"
  exit 1
fi

# Check installation
if [ ! -e "$(which resticprofile)" ] || ! resticprofile version >/dev/null ; then
  echo "resticprofile not found or not executable"
  exit 1
fi

# Generate default-repo secret if missing
if [ ! -f "${SECRET_FILE}" ] ; then
  echo "Generating ${SECRET_FILE}"
  if resticprofile random-key > "${SECRET_FILE}" ; then
    set_permission "${REPOSITORY_SECRET}"
  else
    exit 1
  fi
fi

# Change scheduler to crond when systemd is missing (Alpine)
if [ ! -d "${ROOT_PATH}/etc/systemd/" ] ; then
  sed -iE 's/#scheduler = "systemd"/scheduler = "crond"/g' "${CONFIG_PATH}/profiles.conf.dist" \
    || exit 1
fi

# Generate bash completions
for completion_path in ${COMPLETION_DIRS} ; do
  completion_file="${completion_path}/resticprofile"

  if [ -d "${completion_path}" ] ; then
    echo "Generating ${completion_file}"
    if [ -f "${completion_file}" ] ; then
      chmod u+w "${completion_file}"
    fi
    resticprofile completion-script --bash > "${completion_file}" \
      && chown root:root "${completion_file}" \
      && chmod 0555 "${completion_file}"

    break # install only in the first path
  fi
done

# Merge configuration updates with existing files
if [ -e "$(which diff3)" ] ; then
  echo "Merging updates to config files"

  # Extract previous .dist files from .dist.cache
  if [ -s "${CONFIG_CACHE_FILE}" ] ; then
    cd "${CACHE_PATH}" \
      && tar -xzf "${CONFIG_CACHE_FILE}"
  fi

  new_dist_cache_list="${TEMP_PATH}/dist_cache.list"

  cd "${CONFIG_PATH}" || exit 1

  # Merge existing config with updates from .dist files
  for file in $MERGEABLES ; do
    target_file="${file}"
    new_file="${file}.dist"
    cached_file="${CACHE_PATH}/${new_file}"
    output="${TEMP_PATH}/merged.conf"

    if [ -e "${new_file}" ] ; then
      echo "${new_file}" >> "${new_dist_cache_list}"
    fi

    if [ -e "${target_file}" ] && [ -e "${new_file}" ] && [ -e "${cached_file}" ] ; then

      diff3 --easy-only --merge "${new_file}" "${cached_file}" "${target_file}" > "${output}"

      if [ "$?" = "0" ] || [ "$?" = "1" ] ; then
        if [ "$?" = "1" ] ; then
          backup="${target_file}.prev"
          cp -f "${target_file}" "${backup}" \
            && set_permission "${backup}"
          echo "Conflicts found in \"${target_file}\", please verify. Created ${backup}"
        fi
        if [ -s "${output}" ] ; then
          mv -f "${output}" "${target_file}" \
            && set_permission "${target_file}"
        fi
      else
        echo "Failed merging \"${new_file}\" \"${cached_file}\" \"${target_file}\""
      fi
    fi
  done

  # Create new .dist.cache from current .dist files
  if [ -s "${new_dist_cache_list}" ] ; then
    tar -c --files-from "${new_dist_cache_list}" -zf "${CONFIG_CACHE_FILE}"
  fi
fi

# Unwrap remaining dist files where target file does not exist already
cd "${CONFIG_PATH}" || exit 1

for file in *.dist conf.d/*.dist profiles.d/*.dist repositories.d/*.dist templates/*.dist ; do
  target_file="$(dirname "${file}")/$(basename -s ".dist" "${file}")"

  if [ -e "${target_file}" ] ; then
    rm "${file}"
  else
    mv -f "${file}" "${target_file}" \
      && set_permission "${target_file}"
  fi
done
