package restmapper

import (
	"reflect"
	"sync"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	discoveryfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	coretesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var fakeResources = []*metav1.APIResourceList{
	{
		GroupVersion: appsv1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "deployments", Namespaced: true, Kind: "Deployment"},
		},
	},
	{
		GroupVersion: schema.GroupVersion{
			Group:   "",
			Version: "v1",
		}.String(),
		APIResources: []metav1.APIResource{
			{Name: "pods", Namespaced: true, Kind: "Pod"},
		},
	},
	{
		GroupVersion: schema.GroupVersion{
			Group:   "",
			Version: "v2",
		}.String(),
		APIResources: []metav1.APIResource{
			{Name: "pods", Namespaced: true, Kind: "Pod"},
		},
	},
	{
		GroupVersion: schema.GroupVersion{
			Group:   "extensions",
			Version: "v1beta",
		}.String(),
		APIResources: []metav1.APIResource{
			{Name: "jobs", Namespaced: true, Kind: "Job"},
		},
	},
}

// sum = 10 repeat = 6
var kindTCs = []struct {
	want  schema.GroupVersionResource
	input schema.GroupVersionKind
}{
	{
		want: schema.GroupVersionResource{
			Version:  "v1",
			Resource: "pods",
		},
		input: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "Pod",
		},
	},
	{
		want: schema.GroupVersionResource{
			Version:  "v2",
			Resource: "pods",
		},
		input: schema.GroupVersionKind{
			Version: "v2",
			Kind:    "Pod",
		},
	},
	{
		want: schema.GroupVersionResource{
			Group:    "extensions",
			Version:  "v1beta",
			Resource: "jobs",
		},
		input: schema.GroupVersionKind{
			Group:   "extensions",
			Version: "v1beta",
			Kind:    "Job",
		},
	},
	{
		want: schema.GroupVersionResource{
			Version:  "v1",
			Resource: "pods",
		},
		input: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "Pod",
		},
	},
	{
		want: schema.GroupVersionResource{
			Version:  "v1",
			Resource: "pods",
		},
		input: schema.GroupVersionKind{
			Version: "v1",
			Kind:    "Pod",
		},
	},
	{
		want: schema.GroupVersionResource{
			Version:  "v2",
			Resource: "pods",
		},
		input: schema.GroupVersionKind{
			Version: "v2",
			Kind:    "Pod",
		},
	},
	{
		want: schema.GroupVersionResource{
			Version:  "v2",
			Resource: "pods",
		},
		input: schema.GroupVersionKind{
			Version: "v2",
			Kind:    "Pod",
		},
	},
	{
		want: schema.GroupVersionResource{
			Group:    "extensions",
			Version:  "v1beta",
			Resource: "jobs",
		},
		input: schema.GroupVersionKind{
			Group:   "extensions",
			Version: "v1beta",
			Kind:    "Job",
		},
	},
	{
		want: schema.GroupVersionResource{
			Group:    "extensions",
			Version:  "v1beta",
			Resource: "jobs",
		},
		input: schema.GroupVersionKind{
			Group:   "extensions",
			Version: "v1beta",
			Kind:    "Job",
		},
	},
	{
		want: schema.GroupVersionResource{},
		input: schema.GroupVersionKind{
			Group:   "non-existence",
			Version: "non-existence",
			Kind:    "non-existence",
		},
	},
}

var discoveryClient = &discoveryfake.FakeDiscovery{Fake: &coretesting.Fake{Resources: fakeResources}}

func BenchmarkGetGroupVersionResource(b *testing.B) {
	var option = apiutil.WithCustomMapper(func() (meta.RESTMapper, error) {
		groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
		if err != nil {
			return nil, err
		}
		return restmapper.NewDiscoveryRESTMapper(groupResources), nil
	})

	mapper, err := apiutil.NewDynamicRESTMapper(&rest.Config{}, option)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i += len(kindTCs) {
		for _, tc := range kindTCs {
			_, err := GetGroupVersionResource(mapper, tc.input)
			if err != nil && !meta.IsNoMatchError(err) {
				b.Errorf("GetGroupVersionResource For %v Error:%v", tc.input, err)
			}
		}
	}
}

func BenchmarkGetGroupVersionResourceWithCache(b *testing.B) {
	cachedmapper := &cachedRESTMapper{}

	var option = apiutil.WithCustomMapper(func() (meta.RESTMapper, error) {
		groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
		if err != nil {
			return nil, err
		}
		cachedmapper.gvkToGVR = sync.Map{}
		return restmapper.NewDiscoveryRESTMapper(groupResources), nil
	})

	mapper, err := apiutil.NewDynamicRESTMapper(&rest.Config{}, option)
	if err != nil {
		b.Error(err)
	}

	cachedmapper.restMapper = mapper

	for i := 0; i < b.N; i += len(kindTCs) {
		for _, tc := range kindTCs {
			_, err := GetGroupVersionResource(cachedmapper, tc.input)
			if err != nil && !meta.IsNoMatchError(err) {
				b.Errorf("GetGroupVersionResource For %v Error:%v", tc.input, err)
			}
		}
	}
}

func TestGetGroupVersionResourceWithCache(t *testing.T) {
	cachedmapper := &cachedRESTMapper{}

	var option = apiutil.WithCustomMapper(func() (meta.RESTMapper, error) {
		groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
		if err != nil {
			return nil, err
		}
		cachedmapper.gvkToGVR = sync.Map{}
		return restmapper.NewDiscoveryRESTMapper(groupResources), nil
	})

	mapper, err := apiutil.NewDynamicRESTMapper(&rest.Config{}, option)
	if err != nil {
		t.Error(err)
	}

	cachedmapper.restMapper = mapper

	//one test
	for _, tc := range kindTCs[:9] {
		got, err := GetGroupVersionResource(cachedmapper, tc.input)
		if err != nil {
			t.Errorf("GetGroupVersionResource (%#v) unexpected error: %v", tc.input, err)
			continue
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("GetGroupVersionResource(%#v) = %#v, want %#v", tc.input, got, tc.want)
		}
	}

	//invalidate cache
	_, err = GetGroupVersionResource(cachedmapper, kindTCs[9].input)
	if !meta.IsNoMatchError(err) {
		t.Errorf("GetGroupVersionResource (%#v) unexpected error: %v", kindTCs[9].input, err)
	}

	//one more test
	for _, tc := range kindTCs[:9] {
		got, err := GetGroupVersionResource(cachedmapper, tc.input)
		if err != nil {
			t.Errorf("GetGroupVersionResource (%#v) unexpected error: %v", tc.input, err)
			continue
		}

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("GetGroupVersionResource(%#v) = %#v, want %#v", tc.input, got, tc.want)
		}
	}
}
