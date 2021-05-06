package reconciler

import (
	"context"
	"encoding/json"
	"fmt"

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

type MoneroNodeSetReconciler struct {
	Log    logr.Logger
	Client client.Client
}

func (r *MoneroNodeSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nodeSet, err := r.GetMoneroNodeSet(ctx, req.Name, req.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return EmptyResult(), nil
		}

		return EmptyResult(), fmt.Errorf("get moneronodeset: %w", err)
	}

	nodeSet.ApplyDefaults()

	err = r.ReconcileMoneroNodeSet(ctx, nodeSet)
	if err != nil {
		return EmptyResult(), fmt.Errorf("reconcile moneronodeset: %w", err)
	}

	return EmptyResult(), nil
}

func (r *MoneroNodeSetReconciler) ReconcileMoneroNodeSet(
	ctx context.Context,
	nodeSet *v1alpha1.MoneroNodeSet,
) error {
	objs, err := r.GenerateObjects(ctx, nodeSet)
	if err != nil {
		return fmt.Errorf("setup objs: %w", err)
	}

	if err := r.ApplyObjects(ctx, nodeSet, objs); err != nil {
		return fmt.Errorf("apply objects: %w", err)
	}

	nodeSet.Status.Conditions = []metav1.Condition{
		{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "Succeeded",
			Message:            "objects successfully applied",
		},
	}

	if err := r.Client.Status().Update(ctx, nodeSet); err != nil {
		return fmt.Errorf("status update: %w", err)
	}

	return nil
}

func (r *MoneroNodeSetReconciler) GenerateObjects(
	ctx context.Context,
	nodeSet *v1alpha1.MoneroNodeSet,
) ([]client.Object, error) {
	objs := []client.Object{}

	if nodeSet.Spec.Tor.Enabled {
		hiddenServiceSecret := NewTorHiddenServiceSecret(nodeSet)
		torSecretsRec := &TorSecretsReconciler{}

		if err := torSecretsRec.FillSecret(hiddenServiceSecret); err != nil {
			return nil, fmt.Errorf("fill secret: %w", err)
		}

		hostname, found := hiddenServiceSecret.Data["hostname"]
		if !found {
			return nil, fmt.Errorf("tor hidden service secret '%s' should be filled but isn't - didn't find hostname",
				hiddenServiceSecret.GetName(),
			)
		}
		nodeSet.Status.Tor.Address = string(hostname)

		objs = append(objs,
			NewTorHiddenServiceService(nodeSet),
			NewTorHiddenServiceDeployment(nodeSet),
			NewTorProxyConfigMap(nodeSet),
			NewTorHiddenServiceConfigMap(nodeSet),
			hiddenServiceSecret,
		)
	}

	objs = append(objs,
		NewMoneroService(nodeSet),
		NewMoneroStatefulSet(nodeSet),
	)

	return objs, nil
}

func (r *MoneroNodeSetReconciler) ApplyObjects(
	ctx context.Context,
	nodeSet *v1alpha1.MoneroNodeSet,
	objs []client.Object,
) error {
	for _, o := range objs {
		r.SetOwnerRef(nodeSet, o)

		if err := r.Apply(ctx, o); err != nil {
			return fmt.Errorf("apply '%s %s': %w",
				o.GetObjectKind().GroupVersionKind().String(),
				o.GetName(),
				err,
			)
		}
	}

	return nil
}

func (r *MoneroNodeSetReconciler) GetMoneroNodeSet(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.MoneroNodeSet, error) {
	obj := &v1alpha1.MoneroNodeSet{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, obj); err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", namespace, name, err)
	}

	return obj, nil
}

func (r *MoneroNodeSetReconciler) SetOwnerRef(
	parent *v1alpha1.MoneroNodeSet,
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

func (r *MoneroNodeSetReconciler) Apply(
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

func EmptyResult() ctrl.Result {
	return ctrl.Result{}
}
