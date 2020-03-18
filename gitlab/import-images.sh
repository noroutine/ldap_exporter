#!/bin/bash -eu

get_registry() {
    local image=$1
    [[ -z ${image} ]] && return
    if [[ ${image} =~ ^([^\/]+\.[^\/]+)\/ ]]; then
        echo -n ${BASH_REMATCH[1]}
    else
        echo -n docker.io
    fi
}

get_tag() {
    local image=$1
    [[ -z ${image} ]] && echo -n "latest" && return
    if [[ ${image} =~ :([^\/]+) ]]; then
        echo -n ${BASH_REMATCH[1]}
    else
        echo -n "latest"
    fi
}

get_name() {
    local image=$1
    [[ -z ${image} ]] && return
    if [[ ${image} =~ ([^\/]+\.[^\/]+\/)?([^:]+)(:.+)?$ ]]; then
        echo -n ${BASH_REMATCH[2]}
    fi
}

migrate_image() {
    local image=$1
    [[ -z ${image} ]] && return
    local source_registry=$(get_registry ${image})
    local target_registry=${2}
    [[ -z ${target_registry} ]] && return
    local image_name=$(get_name ${image})
    local image_tag=$(get_tag ${image})

    docker pull ${source_registry}/${image_name}:${image_tag}
    docker tag ${source_registry}/${image_name}:${image_tag} ${target_registry}/${image_name}:${image_tag}
    docker push ${target_registry}/${image_name}:${image_tag}
}

line() {
    local width=40
    # if [[ ! -z $TERM ]]; then
    #     width=$(tput cols)
    # fi

    eval printf '%.0s-' {1..${width}}
}

import_images() {
    declare -a images=(
        "golang:1.14-buster"
    )

    local target_registry=${1:-${DOCKER_HUB:-nexus.noroutine.me:5000}}

    for image in ${images[@]}; do
        printf "\nMigrating %s/%s:%s to %s\n`line`\n" \
            $(get_registry $image) $(get_name $image) $(get_tag $image) \
            ${target_registry}
        migrate_image ${image} ${target_registry}
    done
}

_main() {
    local source_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
    source ${source_dir}/env.sh
    
    import_images
}

[[ "${BASH_SOURCE[0]}" != "${0}" ]] || _main "${@:-}"

# End of file