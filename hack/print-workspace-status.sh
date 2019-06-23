#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export KUBEVIRT_DIR=$(dirname "${BASH_SOURCE}")/..

source "${KUBEVIRT_DIR}/hack/version.sh"
kubevirt::version::get_version_vars

# Prefix with STABLE_ so that these values are saved to stable-status.txt
# instead of volatile-status.txt.
# Stamped rules will be retriggered by changes to stable-status.txt, but not by
# changes to volatile-status.txt.
# IMPORTANT: the camelCase vars should match the lists in hack/version.sh
# and pkg/version/def.bzl.
cat <<EOF
STABLE_BUILD_GIT_COMMIT ${KUBEVIRT_GIT_COMMIT-}
STABLE_BUILD_SCM_STATUS ${KUBEVIRT_GIT_TREE_STATE-}
STABLE_BUILD_SCM_REVISION ${KUBEVIRT_GIT_VERSION-}
STABLE_DOCKER_TAG ${KUBEVIRT_GIT_VERSION/+/_}
gitCommit ${KUBEVIRT_GIT_COMMIT-}
gitTreeState ${KUBEVIRT_GIT_TREE_STATE-}
gitVersion ${KUBEVIRT_GIT_VERSION-}
buildDate $(date \
    ${SOURCE_DATE_EPOCH:+"--date=@${SOURCE_DATE_EPOCH}"} \
    -u +'%Y-%m-%dT%H:%M:%SZ')
EOF
