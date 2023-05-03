#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-"fast"}

kind delete cluster --name "${KIND_CLUSTER_NAME}"