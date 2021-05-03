# monero-operator

A Kubernetes-native way of deploying [Monero] nodes, networks, and miners:
express your intention and let Kubernetes run it for you.


> **you:** _"Hi, I'd like two public nodes, and three miners please"._
>
> **k8s**: _"Sure thing"_
>
> **k8s**: _"It looks like you want two public nodes, but I see 0 running - let me create them for you."_
>
> **k8s**: _"It looks like you want three miners, but I see 0 running - let me create them for you."_
>
> **you:** _"Actually, I changed my mind - I don't want to mine on `minexmr`, I want `cryptonode.social`"._
>
> **k8s**: _"Good choice, pool concentration sucks - let me update your miners for you :)"_



See [./docs](./docs) for detailed documentation about each resource.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Example](#example)
  - [Full node](#full-node)
  - [Mining cluster](#mining-cluster)
  - [Private network](#private-network)
- [Usage](#usage)
- [Support this project](#support-this-project)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->


## Example

### Full node

To run a single full node, all you need to do is create a single-replica `MoneroNodeSet`.

```yaml
kind: MoneroNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: full-node
spec:
  replicas: 1
```

With a MoneroNodeSet you express the intention of having a set of `monerod`
nodes running with a particular configuration: you express your intent, and
`monero-operator` takes care of making it happen.

For instance, with [kubectl-tree](https://github.com/ahmetb/kubectl-tree) we
can see that the operator took care of instantiating a
[StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)


```console
$ kubectl  tree moneronodeset.utxo.com.br full-node
NAMESPACE  NAME                                         READY
default    MoneroNodeSet/full-node                      True 
default    ├─Service/full-node                          -    
default    │ └─EndpointSlice/full-node-d2crv            -    
default    └─StatefulSet/full-node                      -    
default      ├─ControllerRevision/full-node-856644d54d  -    
default      └─Pod/full-node-0                          True 
```

with a pre-configured set of flags 


```console
$ kubectl get pod full-node-0 -ojsonpath={.spec.containers[*].command} | jq '.'
[
  "monerod",
  "--data-dir=/data",
  "--log-file=/dev/stdout",
  "--no-igd",
  "--no-zmq",
  "--non-interactive",
  "--p2p-bind-ip=0.0.0.0",
  "--p2p-bind-port=18080",
  "--rpc-restricted-bind-ip=0.0.0.0",
  "--rpc-restricted-bind-port=18089"
]
```

and a [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) attached with enough disk space for it.

```
$ kubectl get pvc
kubectl get pvc
NAME               STATUS   VOLUME                                     CAPACITY
data-full-node-0   Bound    pvc-1c60e835-d5f9-41c9-8509-b0e4b3b71f6b   200Gi    
```

With that all being declarative, updating our node is a matter of expressing
our new intent by updating the `MoneroNodeSet` definition, and letting the
operator take care of updating things behind the scene.

For instance, assuming we want to now make it public, accepting lots of peers,
having a higher bandwidth limit and using dns blocklist and checkpointing, we'd
patch the object with the following:


```yaml
kind: MoneroNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: full-set
spec:
  replicas: 1

  monerod:
    args:
      - --public
      - --enable-dns-blocklist
      - --enforce-dns-checkpointing
      - --out-peers=1024
      - --in-peers=1024
      - --limit-rate=128000
```

Which would then lead to an update to the node (Kubernetes takes care of
signalling the `monerod`, waiting for it to finish gracefully - did I mention
that it has properly set readiness probes too? -, detaching the disk, etc etc).


### Mining cluster

Similar to `MoneroNodeSet`, with a `MoneroMiningNodeSet` you express the
intention of having a cluster o _x_ replicas running, and then the operator
takes care of making that happen.

For instance, to run a set of 5 miners spread across different Kubernetes
nodes:

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

_ps.: `$(id)` is indeed a thing - wherever you put it, it's going to be interpolated with an integer that identifies the replica_
_pps.: `xmrig` is used under the hood_


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
