#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"
TMP_DIR=$( mktemp -d -t dapr-client-gen-XXXXXXXX )

mkdir -p "${TMP_DIR}/client"
mkdir -p "${PROJECT_ROOT}/pkg/client/dapr"

"${PROJECT_ROOT}"/bin/applyconfiguration-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/dapr/v1alpha1 \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/applyconfiguration

"${PROJECT_ROOT}"/bin/client-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input=dapr/v1alpha1 \
  --clientset-name "versioned"  \
  --input-base=github.com/lburgazzoli/dapr-operator-ng/api \
  --apply-configuration-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/applyconfiguration \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/clientset

"${PROJECT_ROOT}"/bin/lister-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/dapr/v1alpha1 \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/listers

"${PROJECT_ROOT}"/bin/informer-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-base="${TMP_DIR}/client" \
  --input-dirs=github.com/lburgazzoli/dapr-operator-ng/api/dapr/v1alpha1 \
  --versioned-clientset-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/clientset/versioned \
  --listers-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/listers \
  --output-package=github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/informers


cp -R "${TMP_DIR}"/client/github.com/lburgazzoli/dapr-operator-ng/pkg/client/dapr/* "${PROJECT_ROOT}"/pkg/client/dapr