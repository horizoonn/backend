#!/usr/bin/env bash

set -euo pipefail

readonly deploy_dir="${DEPLOY_DIR:-/opt/avito-recap}"
readonly current_image_env="${deploy_dir}/current-image.env"
readonly previous_image_env="${deploy_dir}/previous-image.env"

if [[ ! -f "${previous_image_env}" ]]; then
  echo "previous image tag is unavailable" >&2
  exit 2
fi

if [[ ! -f "${current_image_env}" ]]; then
  echo "current image tag is unavailable" >&2
  exit 3
fi

previous_tag="$(sed -n 's/^IMAGE_TAG=//p' "${previous_image_env}")"
if [[ ! "${previous_tag}" =~ ^[0-9a-f]{40}$ ]]; then
  echo "previous image tag is invalid" >&2
  exit 4
fi

if [[ ! -f "${deploy_dir}/releases/${previous_tag}/deploy-release.sh" ]]; then
  echo "previous release bundle is unavailable" >&2
  exit 5
fi

exec bash "${deploy_dir}/releases/${previous_tag}/deploy-release.sh" "${previous_tag}"
