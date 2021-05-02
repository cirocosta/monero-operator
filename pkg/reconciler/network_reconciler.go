package reconciler

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
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
	nodeSet *v1alpha1.MoneroNetwork,
) error {
	return nil
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
