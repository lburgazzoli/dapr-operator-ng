/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/lburgazzoli/dapr-operator-ng/api/tools/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DaprLister helps list Daprs.
// All objects returned here must be treated as read-only.
type DaprLister interface {
	// List lists all Daprs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Dapr, err error)
	// Daprs returns an object that can list and get Daprs.
	Daprs(namespace string) DaprNamespaceLister
	DaprListerExpansion
}

// daprLister implements the DaprLister interface.
type daprLister struct {
	indexer cache.Indexer
}

// NewDaprLister returns a new DaprLister.
func NewDaprLister(indexer cache.Indexer) DaprLister {
	return &daprLister{indexer: indexer}
}

// List lists all Daprs in the indexer.
func (s *daprLister) List(selector labels.Selector) (ret []*v1alpha1.Dapr, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Dapr))
	})
	return ret, err
}

// Daprs returns an object that can list and get Daprs.
func (s *daprLister) Daprs(namespace string) DaprNamespaceLister {
	return daprNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DaprNamespaceLister helps list and get Daprs.
// All objects returned here must be treated as read-only.
type DaprNamespaceLister interface {
	// List lists all Daprs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Dapr, err error)
	// Get retrieves the Dapr from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Dapr, error)
	DaprNamespaceListerExpansion
}

// daprNamespaceLister implements the DaprNamespaceLister
// interface.
type daprNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Daprs in the indexer for a given namespace.
func (s daprNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Dapr, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Dapr))
	})
	return ret, err
}

// Get retrieves the Dapr from the indexer for a given namespace and name.
func (s daprNamespaceLister) Get(name string) (*v1alpha1.Dapr, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("dapr"), name)
	}
	return obj.(*v1alpha1.Dapr), nil
}