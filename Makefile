build:
	go build -o cmd/controller/controller -v ./cmd/controller


run: build
	./cmd/controller/controller


monerod-image:
	docker build \
		-t utxobr/monerod:v0.17.2.0 \
		-f ./images/monerod.Dockerfile \
			./images
	docker push utxobr/monerod:v0.17.2.0


install-crds:
	kapp deploy -a monero -f ./config/bases/crds.yaml


release:
	KO_DOCKER_REPO=utxobr ko resolve -f ./config/bases > ./config/release.yaml


generate:
	controller-gen \
		crd \
		paths=./pkg/apis/utxo.com.br/v1alpha1 \
		output:stdout > ./config/bases/crds.yaml
	controller-gen \
		rbac:roleName=monero-controller \
		paths=./pkg/reconciler \
		output:stdout > ./config/bases/role.yaml
	controller-gen \
		object \
		paths=./pkg/apis/utxo.com.br/v1alpha1

