package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/valyala/fasttemplate"
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

type MoneroMiningNodeSetReconciler struct {
	Log    logr.Logger
	Client client.Client
}

func (r *MoneroMiningNodeSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	miningSet, err := r.GetMoneroMiningNodeSet(ctx, req.Name, req.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			return EmptyResult(), nil
		}

		return EmptyResult(), fmt.Errorf("get moneronodeset: %w", err)
	}

	err = r.ReconcileMoneroMiningNodeSet(ctx, miningSet)
	if err != nil {
		return EmptyResult(), fmt.Errorf("reconcile moneronodeset: %w", err)
	}

	return EmptyResult(), nil
}

func (r *MoneroMiningNodeSetReconciler) ReconcileMoneroMiningNodeSet(
	ctx context.Context,
	miningSet *v1alpha1.MoneroMiningNodeSet,
) error {

	deployments, err := r.AssembleDeployments(miningSet)
	if err != nil {
		return fmt.Errorf("assemble deployments: %w", err)
	}

	for _, deployment := range deployments {
		if err := r.Apply(ctx, deployment); err != nil {
			return fmt.Errorf("apply: %w", err)
		}
	}

	return nil
}

func (r *MoneroMiningNodeSetReconciler) AssembleDeployments(
	miningSet *v1alpha1.MoneroMiningNodeSet,
) ([]*appsv1.Deployment, error) {
	deployments := make([]*appsv1.Deployment, miningSet.Spec.Replicas)

	var err error
	for i := 0; i < int(miningSet.Spec.Replicas); i++ {
		deployments[i], err = r.AssembleMiningDeployment(miningSet, i)
		if err != nil {
			return nil, fmt.Errorf("assemble moneronodeset '%d': %w", i, err)
		}
	}

	return deployments, nil
}

func (r *MoneroMiningNodeSetReconciler) AssembleMiningDeployment(
	miningSet *v1alpha1.MoneroMiningNodeSet,
	idx int,
) (*appsv1.Deployment, error) {
	command := append([]string{
		"xmrig",
	}, miningSet.Spec.Xmrig.Args...)

	for i, arg := range command {
		t := fasttemplate.New(arg, "$(", ")")
		s := t.ExecuteString(map[string]interface{}{
			"id": strconv.Itoa(idx),
		})

		command[i] = s
	}

	container := corev1.Container{
		Name:    "xmrig",
		Image:   miningSet.Spec.Xmrig.Image,
		Command: command,
	}

	o := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      r.DeploymentName(miningSet, idx),
			Namespace: miningSet.Namespace,
		},

		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: AppLabel(miningSet.Name),
			},
			RevisionHistoryLimit: pointer.Int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: AppLabel(miningSet.Name),
				},

				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: pointer.Int64Ptr(60),
					Containers:                    []corev1.Container{container},
				},
			},
		},
	}

	r.SetOwnerRef(miningSet, o)

	return o, nil
}

func (r *MoneroMiningNodeSetReconciler) DeploymentName(
	miningSet *v1alpha1.MoneroMiningNodeSet,
	idx int,
) string {
	return miningSet.Name + "-" + strconv.Itoa(idx)
}

func (r *MoneroMiningNodeSetReconciler) GetMoneroMiningNodeSet(
	ctx context.Context,
	name, namespace string,
) (*v1alpha1.MoneroMiningNodeSet, error) {
	obj := &v1alpha1.MoneroMiningNodeSet{}
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, obj); err != nil {
		return nil, fmt.Errorf("get %s/%s: %w", namespace, name, err)
	}

	return obj, nil
}

func (r *MoneroMiningNodeSetReconciler) Apply(
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

func (r *MoneroMiningNodeSetReconciler) SetOwnerRef(
	parent *v1alpha1.MoneroMiningNodeSet,
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
