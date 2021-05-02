package reconciler

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

type MoneroNodeSetReconciler struct {
	Log    logr.Logger
	Client client.Client
}

func (r *MoneroNodeSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	return ctrl.Result{}, nil
}
