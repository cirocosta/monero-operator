# CRDs

With `monero-operator` you'll have access to three custom resource definitions:

- [`MoneroNodeSet`](#moneronodeset): set of monero nodes that share a similar configuration
- [`MoneroNetwork`](#moneronetwork): set of monero node sets that form a cluster of interconnected nodes
- [`MoneroMiningNodeSet`](#monerominingnodeset): a set of monero mining nodes
  that perform either solo or pooled mining


## MoneroNodeSet

The MoneroNodeSet CRD provides one with the ability of saying _"I want an **x**
number of nodes that looks like this"_, and then having that materializing
behind the scenes.

```

   MoneroNodeSet
        |
        '--- service
        '--- statefulset -- controllerrevision -- {pod1,    ...,   podN}
        '--- configmap                              |               |
                                                    |               |
                                                  mounts          mounts
                                                pvc_1+configmap  pvc_n+configmap
```


Its definition supports the following fields:

- [`apiVersion`][kubernetes-overview] - Specifies the API version, for example
  `tekton.dev/v1beta1`.
- [`kind`][kubernetes-overview] - Identifies this resource object as a `MoneroNode` object.
- [`metadata`][kubernetes-overview] - Specifies metadata that uniquely identifies the
  `MoneroNode` object. For example, a `name`.
- [`spec`][kubernetes-overview] - Specifies the configuration information for
  this `MoneroNode` object. This must include:
  - `replicas` - number of pods to have running _monerod_
  - `hardAntiAffinity` - force pods to land on different underlying machines
  - `monerod` - Specifies the configuration for the
    monero daemon and details like related proxies for non-clearnet usage.
    - `image`: image to use for launching the pod with _monerod_
    - `config`: extra configuration to be passed down to _monerod_. This is a
      free-form map whose values get passed down to the _monerod.conf_ file
      overriding the default configuration (you can find the final
      _monerod.conf_ by looking at the ConfigMap owned by this object).

[kubernetes-overview]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

For instance:

```yaml
kind: MoneroNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: node-set
spec:
  replicas: 5
  monerod:
    image: utxobr/monerod:v0.17.0.2             # override default image
    config:
      limit-rate-up: 128000                     # override the default config
```


## MoneroNetwork

The MoneroNetwork CRD provides one with the ability of saying "I want a network
of inter-connected monero nodes", and then having that materializing behind the
scenes.

```

   MoneroNetwork
        |
        '--- MoneroNodeSet-1 (pointing at set-2 and set-3)
        '--- MoneroNodeSet-2 (pointing at set-1 and set-3)
        '--- MoneroNodeSet-3 (pointing at set-1 and set-2)

```


The difference between `MoneroNodeSet` and `MoneroNetwork` is that in the first
(nodeset) there's no difference in the configuration passed to each one of the
nodes in the set (they're all configured the same). In the case of the network
type, the nodes are configured independently so that each one if aware of each
other.

Having each node connected to each other, we're able to form completely private
networks (great for testing).


Its definition supports the following fields:

- [`apiVersion`][kubernetes-overview] - Specifies the API version, for example
  `tekton.dev/v1beta1`.
- [`kind`][kubernetes-overview] - Identifies this resource object as a `MoneroNode` object.
- [`metadata`][kubernetes-overview] - Specifies metadata that uniquely identifies the
  `MoneroNode` object. For example, a `name`.
- [`spec`][kubernetes-overview] - Specifies the configuration information for
  this `MoneroNode` object. This must include:
  - [`monerod`](#configuring-monerod) - Specifies the configuration for the
    monero daemon and details like related proxies for non-clearnet usage.

For instance:

```yaml
kind: MoneroNetwork
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: testnet
spec:
  replicas: 5
  monerod:
    image: utxobr/monerod:v0.17.0.2             # override default image
    config:
      limit-rate-up: 128000                     # override the default config
      testnet: 1
```


## MoneroMiningNodeSet

The MoneroMiningNodeSet CRD provides one with the ability of saying "I want
**x** nodes whole sole purpose is mining", and then having that materializing
behind the scenes.

```

   MoneroMiningNodeSet
        |
        '--- deployment -- replicaset --  {pod1,    ...,   podN}
        '--- configmap                        |               |
                                              |               |
                                            mounts          mounts
                                          configmap       configmap
```

Its definition supports the following fields:

- [`apiVersion`][kubernetes-overview] - Specifies the API version, for example
  `tekton.dev/v1beta1`.
- [`kind`][kubernetes-overview] - Identifies this resource object as a `MoneroNode` object.
- [`metadata`][kubernetes-overview] - Specifies metadata that uniquely identifies the
  `MoneroNode` object. For example, a `name`.
- [`spec`][kubernetes-overview] - Specifies the configuration information for
  this `MoneroNode` object. This must include:
  - `xmrig` - Specifies the configuration to be passsed for the

For instance,

```yaml
kind: MoneroMiningNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata:
  name: miners
spec:
  replicas: 5
  xmrig:
    image: utxobr/xmrig:v6.12.1
    config:
      cpu: true
      opencl: false
      cuda: false
      pools:
        - url: pool.supportxmr.com:443
          user: 891B5keCnwXN14hA9FoAzGFtaWmcuLjTDT5aRTp65juBLkbNpEhLNfgcBn6aWdGuBqBnSThqMPsGRjWVQadCrhoAT6CnSL3
          pass: $(pod.name)
          keepalive: true
          tls: true
```
