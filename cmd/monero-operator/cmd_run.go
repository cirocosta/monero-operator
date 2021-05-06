package main

import (
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/cirocosta/monero-operator/pkg/reconciler"
)

type RunCommand struct{}

func (c *RunCommand) Execute(_ []string) error {
	scheme := runtime.NewScheme()
	if err := reconciler.AddToScheme(scheme); err != nil {
		return fmt.Errorf("add to scheme: %w", err)
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	if globalOptions.Verbose {
		cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			return NewDumpTransport(rt)
		}
	}

	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
		Scheme:             scheme,
	})
	if err != nil {
		return fmt.Errorf("new manager: %w", err)
	}

	if err := reconciler.RegisterReconcilers(mgr); err != nil {
		return fmt.Errorf("register reconcilers: %w", err)
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		return fmt.Errorf("mgr start: %w", err)
	}

	return nil
}

func init() {
	parser.AddCommand("run",
		"Run Monero Operator",
		"Run Monero Operator",
		&RunCommand{},
	)
}
