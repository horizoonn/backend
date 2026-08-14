#!/usr/bin/env bash

set -euo pipefail

readonly deploy_dir="${DEPLOY_DIR:-/opt/avito-recap}"
readonly image_tag="${1:-}"
readonly release_dir="${deploy_dir}/releases/${image_tag}"
readonly compose_file="${release_dir}/compose.production.yml"
readonly production_env="${deploy_dir}/.env.production"
readonly current_image_env="${deploy_dir}/current-image.env"
readonly previous_image_env="${deploy_dir}/previous-image.env"
readonly lock_file="${deploy_dir}/deployment.lock"

export PRODUCTION_ENV_FILE="${production_env}"

candidate_image_env=""
temporary_current_env=""
deployment_started=false
pre_deploy_tag=""

cleanup() {
  if [[ -n "${candidate_image_env}" ]]; then
    rm -f "${candidate_image_env}"
  fi
  if [[ -n "${temporary_current_env}" ]]; then
    rm -f "${temporary_current_env}"
  fi
}

write_image_env() {
  local destination=$1
  local tag=$2

  printf 'IMAGE_TAG=%s\n' "${tag}" >"${destination}"
}

validate_release() {
  if [[ ! "${image_tag}" =~ ^[0-9a-f]{40}$ ]]; then
    echo "image tag must be a full lowercase Git commit SHA" >&2
    exit 2
  fi

  if [[ ! -f "${production_env}" ]]; then
    echo "production environment file is missing: ${production_env}" >&2
    exit 3
  fi

  if [[ ! -f "${compose_file}" || ! -f "${release_dir}/Caddyfile" ]]; then
    echo "release bundle is incomplete: ${release_dir}" >&2
    exit 4
  fi
}

acquire_deployment_lock() {
  exec 9>"${lock_file}"
  flock --exclusive 9
}

read_current_release() {
  if [[ ! -f "${current_image_env}" ]]; then
    return
  fi

  pre_deploy_tag="$(sed -n 's/^IMAGE_TAG=//p' "${current_image_env}")"
  if [[ ! "${pre_deploy_tag}" =~ ^[0-9a-f]{40}$ ]]; then
    echo "current image tag is invalid" >&2
    exit 5
  fi
}

prepare_candidate() {
  candidate_image_env="$(mktemp "${deploy_dir}/candidate-image.env.XXXXXX")"
  write_image_env "${candidate_image_env}" "${image_tag}"

  compose=(
    docker compose
    --env-file "${production_env}"
    --env-file "${candidate_image_env}"
    --file "${compose_file}"
  )
}

deploy_candidate() {
  "${compose[@]}" config --quiet
  "${compose[@]}" pull --policy always

  deployment_started=true
  "${compose[@]}" up --detach --wait --remove-orphans
}

promote_candidate() {
  if [[ -n "${pre_deploy_tag}" && "${pre_deploy_tag}" != "${image_tag}" ]]; then
    cp "${current_image_env}" "${previous_image_env}"
  fi

  temporary_current_env="$(mktemp "${deploy_dir}/current-image.env.XXXXXX")"
  write_image_env "${temporary_current_env}" "${image_tag}"
  ln -sfn "${release_dir}" "${deploy_dir}/current-release"
  mv "${temporary_current_env}" "${current_image_env}"
  deployment_started=false
}

restore_current_release() {
  local exit_code=$?
  trap - ERR

  if [[ "${deployment_started}" == true && -n "${pre_deploy_tag}" ]]; then
    local current_compose="${deploy_dir}/releases/${pre_deploy_tag}/compose.production.yml"

    if [[ -f "${current_compose}" ]]; then
      echo "deployment failed; restoring ${pre_deploy_tag}" >&2
      local restore_image_env
      restore_image_env="$(mktemp "${deploy_dir}/restore-image.env.XXXXXX")"
      write_image_env "${restore_image_env}" "${pre_deploy_tag}"
      if docker compose \
        --env-file "${production_env}" \
        --env-file "${restore_image_env}" \
        --file "${current_compose}" \
        up --detach --wait --remove-orphans; then
        write_image_env "${current_image_env}" "${pre_deploy_tag}"
        ln -sfn "${deploy_dir}/releases/${pre_deploy_tag}" "${deploy_dir}/current-release"
      else
        echo "automatic restore failed; manual recovery is required" >&2
      fi
      rm -f "${restore_image_env}"
    fi
  elif [[ "${deployment_started}" == true ]]; then
    echo "deployment failed and no previous release is available" >&2
  fi

  exit "${exit_code}"
}

trap cleanup EXIT
trap restore_current_release ERR

validate_release
acquire_deployment_lock
cd "${release_dir}"
read_current_release
prepare_candidate
deploy_candidate
promote_candidate

echo "deployed ${image_tag}"
