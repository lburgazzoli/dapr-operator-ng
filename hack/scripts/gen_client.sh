#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"
TMP_DIR=$( mktemp -d -t dapr-client-gen-XXXXXXXX )

mkdir -p "${TMP_DIR}/client"
mkdir -p "${PROJECT_ROOT}/pkg/client/operator"

"${PROJECT_ROOT}"/bin/applyconfiguration-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/operator/v1alpha1 \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/applyconfiguration

"${PROJECT_ROOT}"/bin/client-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-base=github.com/lburgazzoli/dapr-operator-ng/api \
  --input=operator/v1alpha1 \
  --fake-clientset=false \
  --clientset-name "versioned"  \
  --apply-configuration-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/applyconfiguration \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/clientset

"${PROJECT_ROOT}"/bin/lister-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/operator/v1alpha1 \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/listers

"${PROJECT_ROOT}"/bin/informer-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/operator/v1alpha1 \
  --versioned-clientset-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/clientset/versioned \
  --listers-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/listers \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/informers

# This should not be needed but for some reasons, the applyconfiguration-gen tool
# sets a wrong APIVersion for the Dapr type (operator/v1alpha1 instead of the one with
# the domain operator.dapr.io/v1alpha1).
#
# See: https://github.com/kubernetes/code-generator/issues/150
sed -i \
  's/WithAPIVersion(\"operator\/v1alpha1\")/WithAPIVersion(\"operator.dapr.io\/v1alpha1\")/g' \
  "${TMP_DIR}"/client/github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/applyconfiguration/operator/v1alpha1/daprcontrolplane.go

cp -r \
  "${TMP_DIR}"/client/github.com/lburgazzoli/dapr-operator-ng/pkg/client/operator/* \
  "${PROJECT_ROOT}"/pkg/client/operator

