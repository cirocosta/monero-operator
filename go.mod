module github.com/cirocosta/monero-operator

go 1.16

require (
	github.com/bmizerany/perks v0.0.0-20141205001514-d9a9656a3a4b
	github.com/cirocosta/go-monero v0.0.0-20210522223237-fd03dfddd5ca
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-logr/logr v0.4.0
	github.com/jessevdk/go-flags v1.5.0
	github.com/prometheus/client_golang v1.10.0
	github.com/valyala/fasttemplate v1.2.1
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/tools v0.1.0 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
)

replace github.com/cirocosta/go-monero => ../go-monero
