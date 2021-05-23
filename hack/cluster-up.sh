#!/bin/bash

set -o errexit
set -o nounset

main () {
	start_kind
}

start_kind() {
        cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 30080
        hostPort: 18080
        protocol: TCP
        listenAddress: "0.0.0.0"
      - containerPort: 30089
        hostPort: 18089
        protocol: TCP
        listenAddress: "0.0.0.0"
EOF
}

main "$@"
