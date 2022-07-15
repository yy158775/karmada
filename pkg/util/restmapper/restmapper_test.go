package restmapper

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/restmapper"
)

func BenchmarkGetGroupVersionResource(b *testing.B) {
	var resources = []*restmapper.APIGroupResources{
		{
			Group: metav1.APIGroup{
				Name: "extensions",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1beta"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1beta"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1beta": {
					{Name: "jobs", Namespaced: true, Kind: "Job"},
					{Name: "pods", Namespaced: true, Kind: "Pod"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
					{Version: "v2"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "pods", Namespaced: true, Kind: "Pod"},
				},
				"v2": {
					{Name: "pods", Namespaced: true, Kind: "Pod"},
				},
			},
		},

		// This group tests finding and prioritizing resources that only exist in non-preferred versions
		{
			Group: metav1.APIGroup{
				Name: "unpreferred",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
					{Version: "v2beta1"},
					{Version: "v2alpha1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "broccoli", Namespaced: true, Kind: "Broccoli"},
				},
				"v2beta1": {
					{Name: "broccoli", Namespaced: true, Kind: "Broccoli"},
					{Name: "peas", Namespaced: true, Kind: "Pea"},
				},
				"v2alpha1": {
					{Name: "broccoli", Namespaced: true, Kind: "Broccoli"},
					{Name: "peas", Namespaced: true, Kind: "Pea"},
				},
			},
		},
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(resources)

	kindTCs := []struct {
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
				Group:    "unpreferred",
				Version:  "v2beta1",
				Resource: "peas",
			},
			input: schema.GroupVersionKind{
				Group:   "unpreferred",
				Version: "v2beta1",
				Kind:    "Pea",
			},
		},
		{
			want: schema.GroupVersionResource{
				Group:    "non-exist",
				Version:  "v1",
				Resource: "non-rs",
			},
			input: schema.GroupVersionKind{},
		},
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range kindTCs {
			_, err := GetGroupVersionResource(restMapper, tc.input)
			if err != nil {
				b.Errorf("GetGroupVersionResource For %v Error:%s", tc.input, err)
			}
		}
	}
}

func BenchmarkGetGroupVersionResourceWithMap(b *testing.B) {
	var resources = []*restmapper.APIGroupResources{
		{
			Group: metav1.APIGroup{
				Name: "extensions",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1beta"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1beta"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1beta": {
					{Name: "jobs", Namespaced: true, Kind: "Job"},
					{Name: "pods", Namespaced: true, Kind: "Pod"},
				},
			},
		},
		{
			Group: metav1.APIGroup{
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
					{Version: "v2"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "pods", Namespaced: true, Kind: "Pod"},
				},
				"v2": {
					{Name: "pods", Namespaced: true, Kind: "Pod"},
				},
			},
		},

		// This group tests finding and prioritizing resources that only exist in non-preferred versions
		{
			Group: metav1.APIGroup{
				Name: "unpreferred",
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
					{Version: "v2beta1"},
					{Version: "v2alpha1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "broccoli", Namespaced: true, Kind: "Broccoli"},
				},
				"v2beta1": {
					{Name: "broccoli", Namespaced: true, Kind: "Broccoli"},
					{Name: "peas", Namespaced: true, Kind: "Pea"},
				},
				"v2alpha1": {
					{Name: "broccoli", Namespaced: true, Kind: "Broccoli"},
					{Name: "peas", Namespaced: true, Kind: "Pea"},
				},
			},
		},
	}

	restMapper := restmapper.NewDiscoveryRESTMapper(resources)

	mapper, err := NewGvk2GvrMapRESTMapper(nil, restMapper)

	if err != nil {
		b.Errorf("NewGvk2GvrMapRESTMapper:%s", err.Error())
	}

	kindTCs := []struct {
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
				Group:    "unpreferred",
				Version:  "v2beta1",
				Resource: "peas",
			},
			input: schema.GroupVersionKind{
				Group:   "unpreferred",
				Version: "v2beta1",
				Kind:    "Pea",
			},
		},
		{
			want: schema.GroupVersionResource{
				Group:    "non-exist",
				Version:  "v1",
				Resource: "non-rs",
			},
			input: schema.GroupVersionKind{},
		},
	}

	for i := 0; i < b.N; i++ {
		for _, tc := range kindTCs {
			_, err := GetGroupVersionResource(mapper, tc.input)
			if err != nil {
				b.Errorf("GetGroupVersionResource For %v Error:%s", tc.input, err)
			}
		}
	}
}
