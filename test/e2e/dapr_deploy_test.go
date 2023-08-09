package e2e

import (
	"testing"

	. "github.com/lburgazzoli/dapr-operator-ng/test/support"
)

func TestDaprDeploy(t *testing.T) {

	test := With(t)
	test.T().Parallel()

	_ = test.NewTestNamespace()
}
