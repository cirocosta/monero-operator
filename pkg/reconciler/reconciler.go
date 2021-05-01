package reconciler

import (
	"context"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
)

type MoneroNodeReconciler struct {
	Log    logr.Logger
	Client client.Client
}

func (r *MoneroNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	return ctrl.Result{
		RequeueAfter: 3 * time.Minute,
	}, nil
}
