#!/usr/bin/env bash

set -eu
set -o pipefail

readonly PROGDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly BUILDPACKDIR="$(cd "${PROGDIR}/.." && pwd)"

function main() {
  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done
  printf "${BUILDPACKDIR}"
  mkdir -p "${BUILDPACKDIR}/bin"

  run::build
  cmd::build
}

function usage() {
  cat <<-USAGE
build.sh [OPTIONS]

Builds the buildpack executables.

OPTIONS
  --help  -h  prints the command usage
USAGE
}

function run::build() {
  printf "${BUILDPACKDIR}"
  if [[ -f "${BUILDPACKDIR}/run/main.go" ]]; then
    pushd "${BUILDPACKDIR}/bin" > /dev/null || return
      printf "%s" "Building run... "

      cd ${BUILDPACKDIR}/run
      GOOS=linux \
        go build \
          -ldflags="-s -w" \
          -o "run"

      echo "Success!"

      for name in detect build; do
        printf "%s" "Linking ${name}... "

        ln -sf "run" "${name}"

        echo "Success!"
      done
    popd > /dev/null || return
  fi
}

function cmd::build() {
  if [[ -d "${BUILDPACKDIR}/cmd" ]]; then
    local name
    for src in "${BUILDPACKDIR}"/cmd/*; do
      name="$(basename "${src}")"
      printf name
      printf "${BUILDPACKDIR}"

      printf "%s" "Building ${name}... "

      GOOS="linux" \
        go build \
          -ldflags="-s -w" \
          -o "${BUILDPACKDIR}/bin/${name}" \
            "${src}/main.go"

      echo "Success!"
    done
  fi
}

main "${@:-}"
