#!/usr/bin/env bash
set -e
set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE}")"/.. && pwd)"
CONTROLLER_GEN_REQ_VERSION="v0.9.1-0.20220629131006-1878064c4cdf"
CONTROLLER_GEN_TMP_DIR=$(mktemp -d)
pushd "$CONTROLLER_GEN_TMP_DIR" >/dev/null
  go install sigs.k8s.io/controller-tools/cmd/controller-gen@$CONTROLLER_GEN_REQ_VERSION
popd >/dev/null
rm -rf "$CONTROLLER_GEN_TMP_DIR"
CONTROLLER_GEN=${GOPATH}/bin/controller-gen

export GO111MODULE="on"
export GOFLAGS="-mod=readonly"
export GOPATH="$(mktemp -d)"
export PROJECT_NAME="github.com/dodopizza/stand-schedule-policy-controller"

# Even when modules are enabled, the code-generator tools always write to
# a traditional GOPATH directory, so fake on up to point to the current
# workspace.
mkdir -p "$GOPATH/src/github.com/dodopizza"
ln -s "${SCRIPT_ROOT}" "$GOPATH/src/$PROJECT_NAME"

OUTPUT_PKG=$PROJECT_NAME/pkg/client
APIS_PKG=$PROJECT_NAME/pkg
CLIENTSET_NAME=versioned
CLIENTSET_PKG_NAME=clientset
COMMON_FLAGS="--go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt"

echo "Generating clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}"
go run k8s.io/code-generator/cmd/client-gen \
  --clientset-name "${CLIENTSET_NAME}" \
  --input-base "" \
  --input "${APIS_PKG}/apis/standschedules/v1" \
  --output-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" \
  ${COMMON_FLAGS}

echo "Generating listers at ${OUTPUT_PKG}/listers"
go run k8s.io/code-generator/cmd/lister-gen \
  --input-dirs "${APIS_PKG}/apis/standschedules/v1" \
  --output-package "${OUTPUT_PKG}/listers" \
  ${COMMON_FLAGS}

echo "Generating informers at ${OUTPUT_PKG}/informers"
go run k8s.io/code-generator/cmd/informer-gen \
  --input-dirs "${APIS_PKG}/apis/standschedules/v1" \
  --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
  --listers-package "${OUTPUT_PKG}/listers" \
  --output-package "${OUTPUT_PKG}/informers" \
  ${COMMON_FLAGS}

for VERSION in v1
do
  echo "Generating ${VERSION} register at ${APIS_PKG}/apis/standschedules/${VERSION}"
  go run k8s.io/code-generator/cmd/register-gen \
    --input-dirs "${APIS_PKG}/apis/standschedules/${VERSION}" \
    --output-package "${APIS_PKG}/apis/standschedules/${VERSION}" \
    ${COMMON_FLAGS}

  echo "Generating ${VERSION} deepcopy at ${APIS_PKG}/apis/standschedules/${VERSION}"
  $CONTROLLER_GEN object:headerFile="${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    paths="${APIS_PKG}/apis/standschedules/${VERSION}"
done