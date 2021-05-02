package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/cirocosta/monero-operator/pkg/reconciler"
)

func init() {
	log.SetLogger(zap.New(zap.UseDevMode(true)))
}

var opts = struct {
	Verbose bool `long:"verbose" short:"v" description:"dump all requests"`
}{}

func run() error {
	scheme := runtime.NewScheme()

	if err := reconciler.AddToScheme(scheme); err != nil {
		return fmt.Errorf("add to scheme: %w", err)
	}

	cfg := config.GetConfigOrDie()
	if opts.Verbose {
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

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(1)
	}

	entryLog := log.Log.WithName("entrypoint")
	entryLog.Info("initializing")

	if err := run(); err != nil {
		entryLog.Error(err, "failed to initialize controller")
		return
	}

	entryLog.Info("finished")
}
