package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/cirocosta/monero-operator/pkg/apis/utxo.com.br/v1alpha1"
)

type MoneroNetworkReconciler struct {
	Log    logr.Logger
	Client client.Client
}

func (r *MoneroNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nodeSet, err := r.GetMoneroNetwork(ctx, req.Name, req.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return EmptyResult(), nil
		}

		return EmptyResult(), fmt.Errorf("get moneronodeset: %w", err)
	}

	err = r.ReconcileMoneroNetwork(ctx, nodeSet)
	if err != nil {
		return EmptyResult(), fmt.Errorf("reconcile moneronodeset: %w", err)
	}

	return EmptyResult(), nil
}

func (r *MoneroNetworkReconciler) ReconcileMoneroNetwork(
	ctx context.Context,
	network *v1alpha1.MoneroNetwork,
) error {

	sets, err := r.AssembleSetOfMoneroNodeSets(network)
	if err != nil {
		return fmt.Errorf("assemble set of moneronodesets: %w", err)
	}

	for _, set := range sets {
		if err := r.Apply(ctx, set); err != nil {
			return fmt.Errorf("apply: %w", err)
		}
	}

	// create a set of MoneroNodeSets
	// compare that with a list of current MoneroNodeSets that we own
	// diff
	// reach

	return nil
}

func (r *MoneroNetworkReconciler) AssembleSetOfMoneroNodeSets(
	network *v1alpha1.MoneroNetwork,
) ([]*v1alpha1.MoneroNodeSet, error) {
	nodeSets := make([]*v1alpha1.MoneroNodeSet, network.Spec.Replicas)

	var err error
	for i := 0; i < int(network.Spec.Replicas); i++ {
		nodeSets[i], err = r.AssembleMoneroNodeSet(network, i)
		if err != nil {
			return nil, fmt.Errorf("assemble moneronodeset '%d': %w", i, err)
		}
	}

	return nodeSets, nil
}

func (r *MoneroNetworkReconciler) AssembleMoneroNodeSet(
	network *v1alpha1.MoneroNetwork,
	idx int,
) (*v1alpha1.MoneroNodeSet, error) {
	spec := *network.Spec.Template.Spec.DeepCopy()
	args := spec.Monerod.Args

	for i := 0; i < int(network.Spec.Replicas); i++ {
		if i == idx {
			continue
		}

		args = MergedSlice(args, []string{
			"--add-exclusive-node=" + r.NodeName(network, i),
		})
	}

	spec.Monerod.Args = args

	o := &v1alpha1.MoneroNodeSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MoneroNodeSet",
			APIVersion: v1alpha1.SchemeGroupVersion.Identifier(),
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      r.NodeName(network, idx),
			Namespace: network.Namespace,
		},

		Spec: spec,
	}

	r.SetOwnerRef(network, o)

	return o, nil
}

func (r *MoneroNetworkReconciler) NodeName(
	network *v1alpha1.MoneroNetwork,
	idx int,
) string {
	return network.Name + "-" + strconv.Itoa(idx)
}

func (r *MoneroNetworkReconciler) GetMoneroNetwork(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.MoneroNetwork, error) {
	obj := &v1alpha1.MoneroNetwork{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, obj); err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", namespace, name, err)
	}

	return obj, nil
}

func (r *MoneroNetworkReconciler) SetOwnerRef(
	parent *v1alpha1.MoneroNetwork,
	obj client.Object,
) {
	if len(obj.GetOwnerReferences()) > 0 {
		return
	}

	obj.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         parent.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:               parent.GetObjectKind().GroupVersionKind().Kind,
			Name:               parent.GetName(),
			UID:                parent.GetUID(),
			BlockOwnerDeletion: pointer.BoolPtr(true),
			Controller:         pointer.BoolPtr(true),
		},
	})
}

func (r *MoneroNetworkReconciler) Apply(
	ctx context.Context,
	obj client.Object,
) error {
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}, existing); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("get: %w", err)
		}

		if err := r.Client.Create(ctx, obj); err != nil {
			return fmt.Errorf("create: %w", err)
		}

		return nil
	}

	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	p := client.RawPatch(
		types.ApplyPatchType,
		b,
	)

	obj.SetResourceVersion(existing.GetResourceVersion())
	if err := r.Client.Patch(ctx, obj, p, &client.PatchOptions{
		FieldManager: "controller",
		Force:        pointer.BoolPtr(true),
	}); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	return nil
}
