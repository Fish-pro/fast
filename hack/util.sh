#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

FAST_GO_PACKAGE="github.com/fast-io/fast"

FAST_TARGET_SOURCE=(
  fast-controller-manager=cmd/controller-manager
  fast-agent=cmd/agent
  fast=cmd/fast
)

# This script holds common bash variables and utility functions.

# This function installs a Go tools by 'go install' command.
# Parameters:
#  - $1: package name, such as "sigs.k8s.io/controller-tools/cmd/controller-gen"
#  - $2: package version, such as "v0.8.0"
function util::install_tools() {
	local package="$1"
	local version="$2"
	echo "go install ${package}@${version}"
	GO111MODULE=on go install "${package}"@"${version}"
	GOPATH=$(go env GOPATH | awk -F ':' '{print $1}')
	export PATH=$PATH:$GOPATH/bin
}

function util:host_platform() {
  echo "$(go env GOHOSTOS)/$(go env GOHOSTARCH)"
}

function util::get_target_source() {
  local target=$1
  for s in "${FAST_TARGET_SOURCE[@]}"; do
    if [[ "$s" == ${target}=* ]]; then
      echo "${s##${target}=}"
      return
    fi
  done
}

function util::version_ldflags() {
  # Git information
  GIT_VERSION=$(util::get_version)
  GIT_COMMIT_HASH=$(git rev-parse HEAD)
  if git_status=$(git status --porcelain 2>/dev/null) && [[ -z ${git_status} ]]; then
    GIT_TREESTATE="clean"
  else
    GIT_TREESTATE="dirty"
  fi
  BUILDDATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
  LDFLAGS="-X github.com/fast-io/fast/pkg/version.gitVersion=${GIT_VERSION} \
                        -X github.com/fast-io/fast/pkg/version.gitCommit=${GIT_COMMIT_HASH} \
                        -X github.com/fast-io/fast/pkg/version.gitTreeState=${GIT_TREESTATE} \
                        -X github.com/fast-io/fast/pkg/version.buildDate=${BUILDDATE}"
  echo $LDFLAGS
}

function util::get_version() {
  git describe --tags --dirty
}

function util::wait_pod_ready() {
    local context_name=$1
    local pod_label=$2
    local pod_namespace=$3

    echo "wait the $pod_label ready..."
    set +e
    util::kubectl_with_retry --context="$context_name" wait --for=condition=Ready --timeout=30s pods -l app=${pod_label} -n ${pod_namespace}
    ret=$?
    set -e
    if [ $ret -ne 0 ];then
      echo "kubectl describe info:"
      kubectl --context="$context_name" describe pod -l app=${pod_label} -n ${pod_namespace}
      echo "kubectl logs info:"
      kubectl --context="$context_name" logs -l app=${pod_label} -n ${pod_namespace}
    fi
    return ${ret}
}
