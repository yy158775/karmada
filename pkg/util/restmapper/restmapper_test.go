package restmapper

import (
	"reflect"
	"sync"
	"testing"

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
		GroupVersion: "apps/v1",
		APIResources: []metav1.APIResource{{Name: "deployments", Namespaced: true, Kind: "Deployment"}},
	},
	{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{{Name: "pods", Namespaced: true, Kind: "Pod"}},
	},
	{
		GroupVersion: "v2",
		APIResources: []metav1.APIResource{{Name: "pods", Namespaced: true, Kind: "Pod"}},
	},
	{
		GroupVersion: "extensions/v1beta",
		APIResources: []metav1.APIResource{{Name: "jobs", Namespaced: true, Kind: "Job"}},
	},
}

// getGVRTestCases organize the test cases for GetGroupVersionResource.
// It can be shared by both benchmark and unit test.
var getGVRTestCases = []struct {
	name        string
	inputGVK    schema.GroupVersionKind
	expectedGVR schema.GroupVersionResource
	expectErr   bool
}{
	{
		name:        "v1,Pod cache miss",
		inputGVK:    schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		expectedGVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
		expectErr:   false,
	},
	{
		name:        "v2,Pod cache miss",
		inputGVK:    schema.GroupVersionKind{Group: "", Version: "v2", Kind: "Pod"},
		expectedGVR: schema.GroupVersionResource{Group: "", Version: "v2", Resource: "pods"},
		expectErr:   false,
	},
	{
		name:        "extensions/v1beta,Job cache miss",
		inputGVK:    schema.GroupVersionKind{Group: "extensions", Version: "v1beta", Kind: "Job"},
		expectedGVR: schema.GroupVersionResource{Group: "extensions", Version: "v1beta", Resource: "jobs"},
		expectErr:   false,
	},
	{
		name:        "v1,Pod cache hit once",
		inputGVK:    schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		expectedGVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
		expectErr:   false,
	},
	{
		name:        "v1,Pod cache hit twice",
		inputGVK:    schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		expectedGVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
		expectErr:   false,
	},
	{
		name:        "v2,Pod cache hit once",
		inputGVK:    schema.GroupVersionKind{Group: "", Version: "v2", Kind: "Pod"},
		expectedGVR: schema.GroupVersionResource{Group: "", Version: "v2", Resource: "pods"},
		expectErr:   false,
	},
	{
		name:        "v2,Pod cache hit twice",
		inputGVK:    schema.GroupVersionKind{Group: "", Version: "v2", Kind: "Pod"},
		expectedGVR: schema.GroupVersionResource{Group: "", Version: "v2", Resource: "pods"},
		expectErr:   false,
	},
	{
		name:        "extensions/v1beta,Job cache hit once",
		inputGVK:    schema.GroupVersionKind{Group: "extensions", Version: "v1beta", Kind: "Job"},
		expectedGVR: schema.GroupVersionResource{Group: "extensions", Version: "v1beta", Resource: "jobs"},
		expectErr:   false,
	},
	{
		name:        "extensions/v1beta,Job cache hit twice",
		inputGVK:    schema.GroupVersionKind{Group: "extensions", Version: "v1beta", Kind: "Job"},
		expectedGVR: schema.GroupVersionResource{Group: "extensions", Version: "v1beta", Resource: "jobs"},
		expectErr:   false,
	},
	{
		name:      "cache miss and invalidate the cache",
		inputGVK:  schema.GroupVersionKind{Group: "non-existence", Version: "non-existence", Kind: "non-existence"},
		expectErr: true,
	},
}

//var kindTCs = []struct {
//	want  schema.GroupVersionResource
//	input schema.GroupVersionKind
//}{
//	{
//		want: schema.GroupVersionResource{
//			Version:  "v1",
//			Resource: "pods",
//		},
//		input: schema.GroupVersionKind{
//			Version: "v1",
//			Kind:    "Pod",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Version:  "v2",
//			Resource: "pods",
//		},
//		input: schema.GroupVersionKind{
//			Version: "v2",
//			Kind:    "Pod",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Group:    "extensions",
//			Version:  "v1beta",
//			Resource: "jobs",
//		},
//		input: schema.GroupVersionKind{
//			Group:   "extensions",
//			Version: "v1beta",
//			Kind:    "Job",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Version:  "v1",
//			Resource: "pods",
//		},
//		input: schema.GroupVersionKind{
//			Version: "v1",
//			Kind:    "Pod",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Version:  "v1",
//			Resource: "pods",
//		},
//		input: schema.GroupVersionKind{
//			Version: "v1",
//			Kind:    "Pod",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Version:  "v2",
//			Resource: "pods",
//		},
//		input: schema.GroupVersionKind{
//			Version: "v2",
//			Kind:    "Pod",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Version:  "v2",
//			Resource: "pods",
//		},
//		input: schema.GroupVersionKind{
//			Version: "v2",
//			Kind:    "Pod",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Group:    "extensions",
//			Version:  "v1beta",
//			Resource: "jobs",
//		},
//		input: schema.GroupVersionKind{
//			Group:   "extensions",
//			Version: "v1beta",
//			Kind:    "Job",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{
//			Group:    "extensions",
//			Version:  "v1beta",
//			Resource: "jobs",
//		},
//		input: schema.GroupVersionKind{
//			Group:   "extensions",
//			Version: "v1beta",
//			Kind:    "Job",
//		},
//	},
//	{
//		want: schema.GroupVersionResource{},
//		input: schema.GroupVersionKind{
//			Group:   "non-existence",
//			Version: "non-existence",
//			Kind:    "non-existence",
//		},
//	},
//}

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

	for i := 0; i < b.N; i += 1 {
		for _, tc := range getGVRTestCases {
			_, err := GetGroupVersionResource(mapper, tc.inputGVK)
			if err != nil && !tc.expectErr {
				b.Errorf("GetGroupVersionResource For %v Error:%v", tc.inputGVK, err)
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

	for i := 0; i < b.N; i += 1 {
		for _, tc := range getGVRTestCases {
			_, err := GetGroupVersionResource(cachedmapper, tc.inputGVK)
			if err != nil && !tc.expectErr {
				b.Errorf("GetGroupVersionResource For %v Error:%v", tc.inputGVK, err)
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

	//test
	for _, tc := range getGVRTestCases[:9] {
		got, err := GetGroupVersionResource(cachedmapper, tc.inputGVK)
		if err != nil && !tc.expectErr {
			t.Errorf("GetGroupVersionResource (%#v) unexpected error: %v", tc.inputGVK, err)
			continue
		}

		if !reflect.DeepEqual(got, tc.expectedGVR) {
			t.Errorf("GetGroupVersionResource(%#v) = %#v, want %#v", tc.inputGVK, got, tc.expectedGVR)
		}
	}

	//invalidate cache
	_, err = GetGroupVersionResource(cachedmapper, getGVRTestCases[9].inputGVK)
	if !meta.IsNoMatchError(err) {
		t.Errorf("GetGroupVersionResource (%#v) unexpected error: %v", getGVRTestCases[9].inputGVK, err)
	}

	//one more test
	for _, tc := range getGVRTestCases[:9] {
		got, err := GetGroupVersionResource(cachedmapper, tc.inputGVK)
		if err != nil {
			t.Errorf("GetGroupVersionResource (%#v) unexpected error: %v", tc.inputGVK, err)
			continue
		}

		if !reflect.DeepEqual(got, tc.expectedGVR) {
			t.Errorf("GetGroupVersionResource(%#v) = %#v, want %#v", tc.inputGVK, got, tc.expectedGVR)
		}
	}
}
