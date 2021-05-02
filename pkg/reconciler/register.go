package reconciler

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1alpha1 "github.com/cirocosta/monero-operator/pkg/apis/utxo.com.br/v1alpha1"
)

func AddToScheme(scheme *runtime.Scheme) error {
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return fmt.Errorf("clientgoscheme addtoscheme: %w", err)
	}

	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("v1alpha1 addtoscheme: %w", err)
	}

	return nil
}

func RegisterReconcilers(mgr manager.Manager) error {
	if err := RegisterMoneroNodeSetReconciler(mgr); err != nil {
		return fmt.Errorf("register moneronodeset reconciler: %w")
	}

	if err := RegisterMoneroNetworkReconciler(mgr); err != nil {
		return fmt.Errorf("register moneronetwork reconciler: %w")
	}

	return nil
}

func RegisterMoneroNetworkReconciler(mgr manager.Manager) error {
	c, err := controller.New("moneronetwork-reconciler", mgr, controller.Options{
		Reconciler: &MoneroNetworkReconciler{
			Log:    mgr.GetLogger().WithName("moneronetwork-reconciler"),
			Client: mgr.GetClient(),
		},
	})
	if err != nil {
		return fmt.Errorf("new controller: %w", err)
	}

	if err := c.Watch(
		&source.Kind{Type: &v1alpha1.MoneroNetwork{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}

func RegisterMoneroNodeSetReconciler(mgr manager.Manager) error {
	c, err := controller.New("moneronodeset-reconciler", mgr, controller.Options{
		Reconciler: &MoneroNodeSetReconciler{
			Log:    mgr.GetLogger().WithName("moneronodeset-reconciler"),
			Client: mgr.GetClient(),
		},
	})
	if err != nil {
		return fmt.Errorf("new controller: %w", err)
	}

	if err := c.Watch(
		&source.Kind{Type: &v1alpha1.MoneroNodeSet{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}
