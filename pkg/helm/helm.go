package helm

import (
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func NewEngine() *Engine {
	return &Engine{
		e:       engine.Engine{},
		decoder: yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme),
	}
}

type Engine struct {
	e       engine.Engine
	decoder runtime.Serializer
}

func (e *Engine) Render(c *chart.Chart, values chartutil.Values) ([]unstructured.Unstructured, error) {
	files, err := engine.Engine{}.Render(c, values)
	if err != nil {
		return nil, errors.Wrap(err, "cannot render a chart")
	}

	result := make([]unstructured.Unstructured, 0)
	for k, v := range files {
		u, err := e.decode([]byte(v))
		if err != nil {
			return nil, errors.Wrapf(err, "cannot decode %s", k)
		}
		if u == nil {
			continue
		}
		if u.GetKind() == "" {
			continue
		}

		result = append(result, *u)
	}

	return result, nil
}

func (e *Engine) decode(content []byte) (*unstructured.Unstructured, error) {
	var obj unstructured.Unstructured

	_, _, err := e.decoder.Decode(content, nil, &obj)
	if err != nil {
		if runtime.IsMissingKind(err) {
			return nil, nil
		}
	}

	return &obj, nil
}
