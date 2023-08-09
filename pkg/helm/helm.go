package helm

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

var decoder = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func Render(c *chart.Chart, values map[string]interface{}) ([]unstructured.Unstructured, error) {
	files, err := engine.Engine{Strict: true}.Render(c, values)
	if err != nil {
		return nil, errors.Wrap(err, "cannot render a chart")
	}

	result := make([]unstructured.Unstructured, 0)
	for k, v := range files {
		u, err := decode([]byte(v))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot decode %s", k)
		}

		result = append(result, u)
	}

	return result, nil
}

func decode(content []byte) (unstructured.Unstructured, error) {
	var obj unstructured.Unstructured

	_, _, err := decoder.Decode(content, nil, &obj)
	if err != nil {
		return obj, errors.Wrap(err, "cannot decode ato unstructured")
	}

	return obj, nil
}
