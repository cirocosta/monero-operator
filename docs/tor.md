# Tor support

## overview

Support for Tor is provided on two fronts:

- facilitating the creation of credentials for hidden services through
  `utxo.com.br/tor`-labelled secrets
- wiring the pods that run `monerod` instances with a Tor sidecar that acts as
  ingress and egress for Tor traffic, as well as applying the proper args for
  `monerod`.

Through the combination of both, one gets the ability of having a full monero
node, on any VPS or cloud provider, serving over both clearnet and tor with no
more than 5 lines of yaml:


```yaml
kind: MoneroNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata: {name: "my-nodes"}
spec:
  tor: {enabled: true}
```

_yes_, powerful.


## secret reconciler

This reconciler, based on labels, is able to take action on those secrets that
would want to be populated with Tor credentials.

All you need to do is create a Secret with the annotation `utxo.com.br/tor: v3`
for a `v3` service (`v2` will be deprecated anyway, so why bother).

_(ps.: if the secret is already popoulated, the reconciler WILL NOT try to
populate it again)_

For instance, we can create a Secret named `tor`


```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tor
  labels:
    utxo.com.br/tor: v3
```

which after reconciliation will see its `data` filled with the content of the
files you'd expect to see under `HiddenServiceDir`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: tor-creds
  annotations:
    utxo.com.br/tor: v3
data:
  hs_ed25519_secret_key: ...
  hs_ed25519_public_key: ...
  hostname: blasblashskasjjha.onion
```

_(you can see if things went good/bad through events emitted by the
reconciler)_

With those filled, we're then able to make use of them in the form of a volume
mount in a Tor sidecar which then directs traffic to the main container's port
through loopback - after all, they're in the same network namespace.

A full example of a highly-available hidden service:


```yaml
---
#
# create an empty but annotated secret that will get populated with the hidden
# service credentials.
#
apiVersion: v1
kind: Secret
metadata:
  name: tor
  annotations:
    utxo.com.br/tor: "v3"

---
#
# fill a ConfigMap with the `torrc` to be loaded by the tor sidecar.
#
apiVersion: v1
kind: ConfigMap
metadata:
  name: tor
data:
  torrc: |-
    HiddenServiceDir /tor-creds
    HiddenServicePort 80 127.0.0.1:80
    HiddenServiceVersion 3

---
#
# the deployment of our application with the application container, as well as
# a sidecar that carries the tor proxy, exposing our app to the tor network.
#
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo
  labels: {app: foo}
spec:
  selector:
    matchLabels: {app: foo}
  template:
    metadata:
      labels: {app: foo}
    spec:
      volumes:
        - name: tor-creds
          secret: {secretName: tor-creds}
        - name: tor-creds
          configMap: {name: torrc}

      containers:
        - image: utxobr/example
          name: my-main-container
          env:
            - name: onion_addr
              valueFrom:
                secretKeyRef:
                  name: tor
                  key: hostname

        - image: utxobr/tor
          name: tor-sidecar
          volumeMounts:
            - name: tor-creds
              mountPath: /tor-creds
            - name: torrc
              mountPath: /torrc
```

_ps.: notice that there's no need for `Service` - that's because we don't need
an external ip or any form of public port; this is a hidden service :)_

an interesting side note here is that we not only are able to expose our
service in the Tor network, but we also have access to it via `socks5` by
making requests to the sidecar under 127.0.0.1:9050 (again, same network
namespace!)

_ps.: note the use of the `ONION_ADDRESS` environment variable - that's in
order to force redeployments to occur whenever there's a change to the secret - see https://ops.tips/notes/kuberntes-secrets/_


## tor-enabled monero nodes

As `MoneroNodeSet`s create plain core Kubernetes resources in order to drive
the execution of `monerod`, we can do the same for enabling Tor support.

Just like with non-Tor nodes, we want to still be able to create notes with
nothing more than a request for monero nodes:

```yaml
kind: MoneroNodeSet
apiVersion: utxo.com.br/v1alpha1
metadata: {name: "my-nodes"}
spec: 
  replicas: 1
```

As Tor support should be just as simple as clearnet, making it Tor-enabled
takes a single line:

```diff
 kind: MoneroNodeSet
 apiVersion: utxo.com.br/v1alpha1
 metadata: {name: "my-nodes"}
 spec: 
   replicas: 1
+  tor: {enabled: true}
```

Under the hood, all that we do then is create one extra primitive, a
`utxo.com.br/tor` labelled `Secret`, which we then mount into a Tor sidecar
container that using the credentials, is then able to proxy traffic from the
Tor network into `monerod` via loopback, as well as serve as a socks5 proxy for
outgoing connections (through loopback as well).


```

        StatefulSet

                ControllerRevision

                        Pod
                                container monerod
                                        -> mounts volume for data
                                        -> points args properly at sidecar

                                containerd torsidecar
                                        -> mounts volume for torrc configmap 
                                        -> mounts volume for hidden svc secrets
                                                -> proxies tor->monerod
                                                -> proxies monerod->tor

```

