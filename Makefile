build:
	go build -o cmd/controller/controller -v ./cmd/controller
run: build
	./cmd/controller/controller
deploy:
	kapp deploy -a monero -f ./config/release.yaml
delete:
	kapp delete -a monero --yes


.PHONY: images
images:
	kbld --images-annotation=false -f ./images/images.yaml > ./images/resolved.yaml


install-crds:
	kapp deploy -a monero -f ./config/bases/crds.yaml --yes
uninstall-crds:
	kapp delete -a monero --yes


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
