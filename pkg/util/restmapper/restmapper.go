package restmapper

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// GetGroupVersionResource is a helper to map GVK(schema.GroupVersionKind) to GVR(schema.GroupVersionResource).
func GetGroupVersionResource(restMapper meta.RESTMapper, gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	restMapping, err := restMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return restMapping.Resource, nil
}

// cachedRESTMapper is a RESTMapper that will provides gvkToGVR cache for RESTMapping Method.
type cachedRESTMapper struct {
	restMapper meta.RESTMapper
	gvkToGVR   sync.Map
}

func (g *cachedRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	//TODO implement me
	return g.restMapper.KindFor(resource)
}

func (g *cachedRESTMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	//TODO implement me
	return g.restMapper.KindsFor(resource)
}

func (g *cachedRESTMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	//TODO implement me
	return g.restMapper.ResourceFor(input)
}

func (g *cachedRESTMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	//TODO implement me
	return g.restMapper.ResourcesFor(input)
}

func (g *cachedRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	//TODO implement me
	return g.restMapper.RESTMappings(gk, versions...)
}

func (g *cachedRESTMapper) ResourceSingularizer(resource string) (singular string, err error) {
	//TODO implement me
	return g.restMapper.ResourceSingularizer(resource)
}

func (g *cachedRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	if len(versions) == 1 {
		currGVK := gk.WithVersion(versions[0])
		value, ok := g.gvkToGVR.Load(currGVK)
		if !ok {
			restMapping, err := g.restMapper.RESTMapping(gk, versions...)
			if err != nil {
				return restMapping, err
			}
			g.gvkToGVR.Store(currGVK, restMapping)
			value = restMapping
		}
		return value.(*meta.RESTMapping), nil
	}

	return g.restMapper.RESTMapping(gk, versions...)
}

// NewcachedRESTMapper returns a cachedRESTMapper for restMapper and cfg.
// The cachedRESTMapper is a RESTMapper that will provides map cache for RESTMapping Method.
func NewcachedRESTMapper(cfg *rest.Config, restMapper meta.RESTMapper) (meta.RESTMapper, error) {
	newmapper := cachedRESTMapper{}

	if restMapper != nil {
		newmapper.restMapper = restMapper
		return &newmapper, nil
	}

	client, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}

	option := apiutil.WithCustomMapper(func() (meta.RESTMapper, error) {
		groupResources, err := restmapper.GetAPIGroupResources(client)
		if err != nil {
			return nil, err
		}
		newmapper.gvkToGVR = sync.Map{}
		return restmapper.NewDiscoveryRESTMapper(groupResources), nil
	})

	restMapper, err = apiutil.NewDynamicRESTMapper(cfg, option)
	if err != nil {
		return nil, err
	}
	newmapper.restMapper = restMapper
	return &newmapper, nil
}
