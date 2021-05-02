package reconciler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	if _, err := r.SetupConfigMap(ctx, nodeSet); err != nil {
		return fmt.Errorf("setup configmap: %w", err)
	}

	if _, err := r.SetupService(ctx, nodeSet); err != nil {
		return fmt.Errorf("setup service: %w", err)
	}

	if _, err := r.SetupStatefulSet(ctx, nodeSet); err != nil {
		return fmt.Errorf("setup statefulset: %w", err)
	}

	return nil
}

func (r *MoneroNodeSetReconciler) SetupConfigMap(
	ctx context.Context,
	nodeSet *v1alpha1.MoneroNodeSet,
) (*corev1.ConfigMap, error) {
	config := NewDefaultMonerodConfig()
	for k, v := range nodeSet.Spec.Monerod.Config {
		config[k] = v
	}

	cm := config.ConfigMap(nodeSet.Name, nodeSet.Namespace)
	r.SetOwnerRef(nodeSet, cm)

	if err := r.Apply(ctx, cm); err != nil {
		return nil, fmt.Errorf("apply: %w", err)
	}

	return nil, nil
}

func (r *MoneroNodeSetReconciler) SetupService(
	ctx context.Context,
	nodeSet *v1alpha1.MoneroNodeSet,
) (*corev1.Service, error) {

	svc := NewService(nodeSet.Name, nodeSet.Namespace)
	r.SetOwnerRef(nodeSet, svc)

	if err := r.Apply(ctx, svc); err != nil {
		return nil, fmt.Errorf("apply: %w", err)
	}

	return svc, nil
}

func (r *MoneroNodeSetReconciler) SetupStatefulSet(
	ctx context.Context,
	nodeSet *v1alpha1.MoneroNodeSet,
) (*appsv1.StatefulSet, error) {

	ss := NewStatefulSet(nodeSet.Name, nodeSet.Namespace)
	r.SetOwnerRef(nodeSet, ss)

	if err := r.Apply(ctx, ss); err != nil {
		return nil, fmt.Errorf("apply: %w", err)
	}

	return ss, nil
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
	nodeSet *v1alpha1.MoneroNodeSet,
	obj client.Object,
) {
	obj.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion:         nodeSet.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:               nodeSet.GetObjectKind().GroupVersionKind().Kind,
			Name:               nodeSet.GetName(),
			UID:                nodeSet.GetUID(),
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
