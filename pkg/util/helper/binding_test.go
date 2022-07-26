package helper

import (
	"errors"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"

	policyv1alpha1 "github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"

	workv1alpha2 "github.com/karmada-io/karmada/pkg/apis/work/v1alpha2"
)

var crbLabelsKeyIndex = "crbLabels"

func TestHasScheduledReplica(t *testing.T) {
	tests := []struct {
		name           string
		scheduleResult []workv1alpha2.TargetCluster
		want           bool
	}{
		{
			name: "all targetCluster have replicas",
			scheduleResult: []workv1alpha2.TargetCluster{
				{
					Name:     "foo",
					Replicas: 1,
				},
				{
					Name:     "bar",
					Replicas: 2,
				},
			},
			want: true,
		},
		{
			name: "a targetCluster has replicas",
			scheduleResult: []workv1alpha2.TargetCluster{
				{
					Name:     "foo",
					Replicas: 1,
				},
				{
					Name: "bar",
				},
			},
			want: true,
		},
		{
			name: "another targetCluster has replicas",
			scheduleResult: []workv1alpha2.TargetCluster{
				{
					Name: "foo",
				},
				{
					Name:     "bar",
					Replicas: 1,
				},
			},
			want: true,
		},
		{
			name: "not assigned replicas for a cluster",
			scheduleResult: []workv1alpha2.TargetCluster{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasScheduledReplica(tt.scheduleResult); got != tt.want {
				t.Errorf("HasScheduledReplica() = %v, want %v", got, tt.want)
			}
		})
	}
}

var fakeResources = []*workv1alpha2.ClusterResourceBinding{
	{ObjectMeta: metav1.ObjectMeta{Name: "one", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "one"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "two", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "two"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "three", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "three"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "four", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "four"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "five", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "five"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "six", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "six"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "seven", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "seven"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "eight", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "eight"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "nine", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "nine"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "ten", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "ten"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "1", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "one"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "2", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "two"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "3", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "three"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "4", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "four"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "5", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "five"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "6", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "six"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "7", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "seven"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "8", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "eight"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "9", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "nine"}}},
	{ObjectMeta: metav1.ObjectMeta{Name: "0", Labels: map[string]string{policyv1alpha1.ClusterPropagationPolicyLabel: "ten"}}},
}

var labelIndexerFunc = func(obj interface{}) ([]string, error) {
	crb, ok := obj.(*workv1alpha2.ClusterResourceBinding)
	if !ok {
		return nil, errors.New("assert failed")
	}
	res := make([]string, 0)
	for key, value := range crb.Labels {
		if key == policyv1alpha1.ClusterPropagationPolicyLabel {
			res = append(res, key+"="+value)
		}
	}
	return res, nil
}

var indexer = cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{crbLabelsKeyIndex: labelIndexerFunc})

func TestGetClusterResourceBindings(t *testing.T) {
	for _, crb := range fakeResources {
		err := indexer.Add(crb)
		if err != nil {
			t.Fatal(err)
		}
	}

	labelSet := labels.Set{
		policyv1alpha1.ClusterPropagationPolicyLabel: "one",
	}

	objs, err := indexer.ByIndex(crbLabelsKeyIndex, labelSet.String())

	if err != nil {
		t.Fatal(err)
	}

	if len(objs) != 2 {
		t.Fatal("sum is not correct")
	}
	obj0 := objs[0].(*workv1alpha2.ClusterResourceBinding)
	obj1 := objs[1].(*workv1alpha2.ClusterResourceBinding)
	if !(obj0.Name == "one" && obj1.Name == "1") && !(obj0.Name == "1" && obj1.Name == "one") {
		t.Error("not match")
	}
}

func BenchmarkGetClusterResourceBindingsWithLabelByIndex(b *testing.B) {
	for _, crb := range fakeResources {
		err := indexer.Add(crb)
		if err != nil {
			b.Fatal(err)
		}
	}

	labelSet := labels.Set{
		policyv1alpha1.ClusterPropagationPolicyLabel: "one",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		indexer.ByIndex(crbLabelsKeyIndex, labelSet.String())
	}
}

func BenchmarkGetClusterResourceBindingsWithLabelByListAll(b *testing.B) {
	for _, crb := range fakeResources {
		err := indexer.Add(crb)
		if err != nil {
			b.Fatal(err)
		}
	}

	labelSet := labels.Set{
		policyv1alpha1.ClusterPropagationPolicyLabel: "one",
	}

	labelSel := labels.SelectorFromSet(labelSet)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		objs := indexer.List()
		results := make([]runtime.Object, 0)
		for _, item := range objs {

			obj, isObj := item.(runtime.Object)

			if !isObj {
				b.Errorf("cache contained %T, which is not an Object", obj)
			}

			meta, err := apimeta.Accessor(obj)
			if err != nil {
				b.Error(err)
			}
			if labelSel != nil {
				lbls := labels.Set(meta.GetLabels())
				if !labelSel.Matches(lbls) {
					continue
				}
			}
			results = append(results, obj)
		}
	}
}
