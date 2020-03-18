#!/usr/bin/env bash -eux

for ci_env_var in "${!BUILD_@}"; do
    export "CI_${ci_env_var#BUILD_}"="${!ci_env_var}"
done