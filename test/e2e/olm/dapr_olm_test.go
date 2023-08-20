package operator

import (
	"github.com/onsi/gomega"
	olmV1 "github.com/operator-framework/api/pkg/operators/v1"
	olmV1Alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/rs/xid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
	"time"

	. "github.com/lburgazzoli/dapr-operator-ng/test/support"
)

func TestDaprDeploy(t *testing.T) {
	test := With(t)

	ns := test.NewTestNamespace()
	id := xid.New().String()

	image := os.Getenv("CATALOG_CONTAINER_IMAGE")

	test.Expect(image).
		ToNot(gomega.BeEmpty())

	_, err := test.Client().OLM.OperatorsV1().OperatorGroups(ns.Name).Create(
		test.Ctx(),
		&olmV1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      id,
				Namespace: ns.Name,
			},
		},
		metav1.CreateOptions{},
	)

	test.Expect(image).
		ToNot(gomega.BeEmpty())

	catalog, err := test.Client().OLM.OperatorsV1alpha1().CatalogSources(ns.Name).Create(
		test.Ctx(),
		&olmV1Alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      id,
				Namespace: ns.Name,
			},
			Spec: olmV1Alpha1.CatalogSourceSpec{
				SourceType:  "grpc",
				Image:       image,
				DisplayName: "Dapr.io Catalog",
				Publisher:   "dapr.io",
				GrpcPodConfig: &olmV1Alpha1.GrpcPodConfig{
					SecurityContextConfig: "restricted",
				},
				UpdateStrategy: &olmV1Alpha1.UpdateStrategy{
					RegistryPoll: &olmV1Alpha1.RegistryPoll{
						Interval: &metav1.Duration{
							Duration: 10 * time.Minute,
						},
					},
				},
			},
		},
		metav1.CreateOptions{},
	)

	test.Expect(err).
		ToNot(gomega.HaveOccurred())

	test.Eventually(CatalogSource(test, catalog.Name, catalog.Namespace), TestTimeoutLong).Should(
		gomega.WithTransform(ExtractCatalogState(), gomega.Equal("READY")),
	)

}
