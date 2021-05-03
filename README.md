# monero-operator

A Kubernetes-native way of deploying [Monero] nodes, networks, and miners:
express your intention and let Kubernetes run it for you.


> **you:** _"Hi, I'd like two public nodes, and three miners please"._

> **k8s**: _"Sure thing"_

> **k8s**: _"It looks like you want two public nodes, but I see 0 running - let me create them for you."_

> **k8s**: _"It looks like you want three miners, but I see 0 running - let me create them for you."_



See [./docs](./docs) for detailed documentation about each resource.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Example](#example)
  - [Full node](#full-node)
  - [Mining cluster](#mining-cluster)
  - [Private network](#private-network)
- [Usage](#usage)
- [Support this project](#support-this-project)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Example

### Full node

Run a single public full node:

```yaml
kind: MoneroNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: node-set
spec:
  replicas: 1
  diskSize: 200Gi

  monerod:
    args:
      - --public
      - --enable-dns-blocklist
      - --enforce-dns-checkpointing
      - --out-peers=1024
      - --in-peers=1024
      - --limit-rate=128000
```


### Mining cluster

Run a set of 5 miners spread across different Kubernetes nodes:

```yaml
kind: MoneroMiningNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: miners
spec:
  replicas: 5
  hardAntiAffinity: true

  xmrig:
    args:
      - -o
      - cryptonote.social:5556
      - -u
      - 891B5keCnwXN14hA9FoAzGFtaWmcuLjTDT5aRTp65juBLkbNpEhLNfgcBn6aWdGuBqBnSThqMPsGRjWVQadCrhoAT6CnSL3.node-$(id)
      - --tls
```


### Private network

```yaml
kind: MoneroNetwork
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: network
spec:
  replicas: 3

  template:
    spec:
      replicas: 1
      monerod:
        args:
          - --regtest
          - --fixed-difficulty=1
```


## Usage

1. install

```bash
# submit the customresourcedefinition objects, deployment, role-base access
# control configs, etc.
#
kubectl apply -f ./config/release.yaml
```

2. submit a description of the intention of having a `monero` node running

```yaml
apiVersion: utxo.com.br/v1alpha1
kind: MoneroNodeSet
spec:
  replicas: 1
  monerod:
    image: utxobr/monerod:v0.17.2
```

(_see [`./docs`](./docs) for more details)_

3. grab the details

```console
$ kubectl get moneronode node-1 -o jsonpath={.status}
```


## Support this project

All you see here has been done during my personal time as a way of helping the Monero community.

Consider donating if you find this helpful or you make use of it :)

![](assets/donate.png)

891B5keCnwXN14hA9FoAzGFtaWmcuLjTDT5aRTp65juBLkbNpEhLNfgcBn6aWdGuBqBnSThqMPsGRjWVQadCrhoAT6CnSL3


[Monero]: https://www.getmonero.org/
