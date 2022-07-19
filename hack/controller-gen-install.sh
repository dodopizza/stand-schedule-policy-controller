#!/usr/bin/env bash
set -e

CONTROLLER_GEN_REQ_VERSION=${1}
CONTROLLER_GEN_TMP_DIR=$(mktemp -d)

cd $CONTROLLER_GEN_TMP_DIR
go install sigs.k8s.io/controller-tools/cmd/controller-gen@$CONTROLLER_GEN_REQ_VERSION
rm -rf $CONTROLLER_GEN_TMP_DIR

CONTROLLER_GEN=${GOPATH}/bin/controller-gen
echo $CONTROLLER_GEN