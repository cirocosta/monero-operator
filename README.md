# monero-operator

A Kubernetes-native way of deploying Monero nodes: express your intention and
let Kubernetes run it for you.

Features:

- start up from a pre-synced chain file
- optional observability stack
- Tor and I2P
- notifications


## Usage

1. install

```console
# submit the customresourcedefinition objects, deployment, role-base access
# control configs, etc.
#
$ kubectl apply -f ./config/release.yaml
```

2. submit a description of the intention of having a `monero` node running

```yaml
apiVersion: utxo.com.br/v1alpha1
kind: MoneroNode
spec:
  monerod:
    image: utxobr/monerod:v0.17.2
```

(_see [`./docs`](./docs) for more details)_

3. grab the details

```console
$ kubectl get moneronode node-1 -o jsonpath={.status}
```

## Support

All you see here has been done during my personal time as a way of helping the Monero community.

Consider donating if you find this helpful or you make use of it :)

![](assets/donate.png)

891B5keCnwXN14hA9FoAzGFtaWmcuLjTDT5aRTp65juBLkbNpEhLNfgcBn6aWdGuBqBnSThqMPsGRjWVQadCrhoAT6CnSL3
