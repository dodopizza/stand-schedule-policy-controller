#!/usr/bin/env bash
set -e
set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE}")"/.. && pwd)"

echo "Install controller-gen"
go install \
  "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.1-0.20220629131006-1878064c4cdf"
export CONTROLLER_GEN=${GOPATH}/bin/controller-gen

# Even when modules are enabled, the code-generator tools always write to
# a traditional GOPATH directory, so fake on up to point to the current
# workspace.

export GO111MODULE="on"
export GOFLAGS="-mod=readonly"
export GOPATH="$(mktemp -d)"
export ORG_NAME="github.com/dodopizza"
export PROJECT_NAME="$ORG_NAME/stand-schedule-policy-controller"

mkdir -p "$GOPATH/src/$ORG_NAME"
ln -s "${SCRIPT_ROOT}" "$GOPATH/src/$PROJECT_NAME"

# Package name with defined CRD types
APIS_PKG=$PROJECT_NAME/pkg/apis/standschedules
APIS_PKG_PRIMARY_VERSION=v1
APIS_PKG_SUPPORTED_VERSIONS=(
  v1
)

# Package name with generated clients (resource client, factories, informers, listers, etc)
CLIENT_PKG=$PROJECT_NAME/pkg/client
CLIENT_CLIENTSET_PKG="${CLIENT_PKG}/clientset"
CLIENT_CLIENTSET_VERSIONED_PKG_NAME=versioned
CLIENT_LISTERS_PKG="${CLIENT_PKG}/listers"
CLIENT_INFORMERS_PKG="${CLIENT_PKG}/informers"

CRDS_PATH="${SCRIPT_ROOT}/crds"

# Flags that shared between various generators
GENERATOR_FLAGS="--go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt"

echo "Generating CRDs at ${CRDS_PATH}"
  $CONTROLLER_GEN crd:crdVersions=v1 \
    paths="${APIS_PKG}/v1" \
    output:crd:artifacts:config="${CRDS_PATH}"
  mv "${CRDS_PATH}/automation.dodois.io_standschedulepolicies.yaml" "${CRDS_PATH}/StandSchedulePolicy.yaml"

echo "Generating ClientSet at ${CLIENT_CLIENTSET_PKG}"
go run k8s.io/code-generator/cmd/client-gen \
  --clientset-name "${CLIENT_CLIENTSET_VERSIONED_PKG_NAME}" \
  --input-base "" \
  --input "${APIS_PKG}/${APIS_PKG_PRIMARY_VERSION}" \
  --output-package "${CLIENT_CLIENTSET_PKG}" \
  ${GENERATOR_FLAGS}

echo "Generating listers at ${CLIENT_LISTERS_PKG}"
go run k8s.io/code-generator/cmd/lister-gen \
  --input-dirs "${APIS_PKG}/${APIS_PKG_PRIMARY_VERSION}" \
  --output-package "${CLIENT_LISTERS_PKG}" \
  ${GENERATOR_FLAGS}

echo "Generating informers at ${CLIENT_INFORMERS_PKG}"
go run k8s.io/code-generator/cmd/informer-gen \
  --input-dirs "${APIS_PKG}/${APIS_PKG_PRIMARY_VERSION}" \
  --versioned-clientset-package "${CLIENT_CLIENTSET_PKG}/${CLIENT_CLIENTSET_VERSIONED_PKG_NAME}" \
  --listers-package "$CLIENT_LISTERS_PKG" \
  --output-package "${CLIENT_INFORMERS_PKG}" \
  ${GENERATOR_FLAGS}

for VERSION in "${APIS_PKG_SUPPORTED_VERSIONS[@]}"
do
  echo "Generating ${VERSION} register at ${APIS_PKG}/${VERSION}"
  go run k8s.io/code-generator/cmd/register-gen \
    --input-dirs "${APIS_PKG}/${VERSION}" \
    --output-package "${APIS_PKG}/${VERSION}" \
    ${GENERATOR_FLAGS}

  echo "Generating ${VERSION} deepcopy at ${APIS_PKG}/${VERSION}"
  $CONTROLLER_GEN object:headerFile="${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    paths="${APIS_PKG}/${VERSION}"
done