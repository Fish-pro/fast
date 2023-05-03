#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail

GOARCH=$(go env GOARCH)
REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
KIND_CLUSTER_NAME=${KIND_CLUSTER_NAME:-"fast"}
source "${REPO_ROOT}"/hack/util.sh

# step1: create cluster by kind
kind create cluster --name "${KIND_CLUSTER_NAME}" --config "${REPO_ROOT}"/artifacts/kind/kind.yaml

# step2: build cni plugins
make fast GOOS="linux"
docker cp "${REPO_ROOT}"/artifacts/cni/node-99-fast.conf "${KIND_CLUSTER_NAME}"-worker:/etc/cni/net.d/99-fast.conf
docker cp "${REPO_ROOT}"/_output/bin/linux/"${GOARCH}"/fast "${KIND_CLUSTER_NAME}"-worker:/opt/cni/bin/fast

docker cp "${REPO_ROOT}"/artifacts/cni/master-99-fast.conf "${KIND_CLUSTER_NAME}"-control-plane:/etc/cni/net.d/99-fast.conf
docker cp "${REPO_ROOT}"/_output/bin/linux/"${GOARCH}"/fast "${KIND_CLUSTER_NAME}"-control-plane:/opt/cni/bin/fast

# step3: build image
export VERSION="latest"
export REGISTRY="fishpro3/fast"
make images GOOS="linux" --directory="${REPO_ROOT}"

# step4: load components images
kind load docker-image "${REGISTRY}/fast-agent:${VERSION}" --name="${KIND_CLUSTER_NAME}"
kind load docker-image "${REGISTRY}/fast-controller-manager:${VERSION}" --name="${KIND_CLUSTER_NAME}"
kind load docker-image nginx:stable --name="${KIND_CLUSTER_NAME}"

# step5: create component
kubectl --context kind-"${KIND_CLUSTER_NAME}" apply -f "${REPO_ROOT}"/artifacts/deploy/namespace.yaml
sleep 5
kubectl --context kind-"${KIND_CLUSTER_NAME}" apply -f "${REPO_ROOT}"/artifacts/deploy/serviceaccount.yaml
kubectl --context kind-"${KIND_CLUSTER_NAME}" apply -f "${REPO_ROOT}"/artifacts/deploy/clusterrolebinding.yaml
kubectl --context kind-"${KIND_CLUSTER_NAME}" apply -f "${REPO_ROOT}"/artifacts/deploy/fast-agent.yaml
kubectl --context kind-"${KIND_CLUSTER_NAME}" apply -f "${REPO_ROOT}"/artifacts/deploy/fast-controller-manager.yaml

# step6: deploy example
kubectl --context kind-"${KIND_CLUSTER_NAME}" apply -f "${REPO_ROOT}"/charts/fast/_crds/bases
sleep 10
kubectl --context kind-"${KIND_CLUSTER_NAME}" apply -f "${REPO_ROOT}"/artifacts/example
