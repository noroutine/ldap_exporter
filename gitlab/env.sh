#!/bin/bash -eux

include_env() {
    local source_dir=${1:-}
    local build_environment=${2:-}
    
    [[ ! -z ${source_dir} ]]
    [[ ! -z ${build_environment} ]]

    export NO_PROXY=${NO_PROXY:-}
    export no_proxy=${NO_PROXY}

    if [[ -d "${source_dir}/env.${build_environment}" ]]; then
      export PROJECT_SOURCE_DIR=$(dirname ${source_dir})
      export BUILD_ENVIRONMENT=${build_environment}
      export DEPLOY_ENVIRONMENT_DIR=${source_dir}/env.${build_environment}

      echo "Export variables for ${BUILD_ENVIRONMENT}"

      for inc in ${source_dir}/env.${build_environment}/*.sh; do
        if [[ -r ${inc} ]]; then
          source ${inc}
        fi
      done
      unset inc
    fi
}

_main() {
    local source_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
    local build_environment=${BUILD_ENVIRONMENT:-default}
    include_env ${source_dir} ${build_environment}
}

_main "${@:-}"

# End of file