package reconciler

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		return fmt.Errorf("register nodeset reconciler: %w", err)
	}

	if err := RegisterMoneroNetworkReconciler(mgr); err != nil {
		return fmt.Errorf("register network reconciler: %w", err)
	}

	if err := RegisterMoneroMiningNodeSetReconciler(mgr); err != nil {
		return fmt.Errorf("register miningnodeset reconciler: %w", err)
	}

	if err := RegisterTorSecretsReconciler(mgr); err != nil {
		return fmt.Errorf("register secrets reconciler: %w", err)
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

func RegisterMoneroMiningNodeSetReconciler(mgr manager.Manager) error {
	c, err := controller.New("monerominingnodeset-reconciler", mgr, controller.Options{
		Reconciler: &MoneroMiningNodeSetReconciler{
			Log:    mgr.GetLogger().WithName("monerominingnodeset-reconciler"),
			Client: mgr.GetClient(),
		},
	})
	if err != nil {
		return fmt.Errorf("new controller: %w", err)
	}

	if err := c.Watch(
		&source.Kind{Type: &v1alpha1.MoneroMiningNodeSet{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	return nil
}

func RegisterTorSecretsReconciler(mgr manager.Manager) error {
	c, err := controller.New("torsecrets-reconciler", mgr, controller.Options{
		Reconciler: &TorSecretsReconciler{
			Log:    mgr.GetLogger().WithName("torsecrets-reconciler"),
			Client: mgr.GetClient(),
		},
	})
	if err != nil {
		return fmt.Errorf("new controller: %w", err)
	}

	p, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"utxo.com.br/tor": "v3",
		},
	})
	if err != nil {
		return fmt.Errorf("predicate: %w", err)
	}

	if err := c.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		&handler.EnqueueRequestForObject{},
		p,
	); err != nil {
		return fmt.Errorf("watch secrets: %w", err)
	}

	return nil
}
