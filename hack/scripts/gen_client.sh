#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"
TMP_DIR=$( mktemp -d -t dapr-client-gen-XXXXXXXX )

mkdir -p "${TMP_DIR}/client"
mkdir -p "${PROJECT_ROOT}/pkg/client/tools"

"${PROJECT_ROOT}"/bin/applyconfiguration-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/tools/v1alpha1 \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/applyconfiguration

"${PROJECT_ROOT}"/bin/client-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-base=github.com/lburgazzoli/dapr-operator-ng/api \
  --input=tools/v1alpha1 \
  --fake-clientset=false \
  --clientset-name "versioned"  \
  --apply-configuration-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/applyconfiguration \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/clientset

"${PROJECT_ROOT}"/bin/lister-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/tools/v1alpha1 \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/listers

"${PROJECT_ROOT}"/bin/informer-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/tools/v1alpha1 \
  --versioned-clientset-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/clientset/versioned \
  --listers-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/listers \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/informers

# This should not be needed but for some reasons, the applyconfiguration-gen tool
# sets a wrong APIVersion for the Dapr type (tools/v1alpha1 instead of the one with
# the domain tools.dapr.io/v1alpha1).
#
# See: https://github.com/kubernetes/code-generator/issues/150
sed -i \
  's/WithAPIVersion(\"tools\/v1alpha1\")/WithAPIVersion(\"tools.dapr.io\/v1alpha1\")/g' \
  "${TMP_DIR}"/client/github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/applyconfiguration/tools/v1alpha1/dapr.go

cp -r \
  "${TMP_DIR}"/client/github.com/lburgazzoli/dapr-operator-ng/pkg/client/tools/* \
  "${PROJECT_ROOT}"/pkg/client/tools

