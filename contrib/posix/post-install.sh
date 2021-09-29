#!/usr/bin/env sh
set -e

CONFIG_PATH="/etc/resticprofile"
SECRET_FILE="${CONFIG_PATH}/conf.d/default-repository.secret"

# Fix permissions (only root may edit and read since password
# & tokens can be in any of the files)
if [ -d "${CONFIG_PATH}" ] ; then
  chown -R root:root "${CONFIG_PATH}"
  chmod -R 0640 "${CONFIG_PATH}"
else
  echo "config path (${CONFIG_PATH}) not found"
  exit 1
fi

# Check installation
if [ ! -e "$(which resticprofile)" ] || ! resticprofile version ; then
  echo "resticprofile not found or not executable"
  exit 1
fi

# Generate default-repo secret (if missing)
if [ ! -f "${SECRET_FILE}" ] ; then
  echo "Generating ${SECRET_FILE}"
  resticprofile random-key > "${SECRET_FILE}"
fi

# Unwrap dist files (if target is missing)
cd "${CONFIG_PATH}"
for file in conf.d/*.dist profiles.d/*.dist templates/*.dist ; do
  target_file="$(dirname "${file}")/$(basename -s ".dist" "${file}")"
  if [ -e "${target_file}" ] ; then
    echo "Skipping ${target_file}. File already exists"
    rm "${file}"
  else
    mv -f "${file}" "${target_file}"
  fi
done
