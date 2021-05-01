# MoneroNode

- [Overview](#overview)
- [Configuring a `MoneroNode`](#configuring-a-moneronode)


## Overview

The MoneroNode CRD provides one with the ability of saying "I want a node that
looks like this", and then having that materializing behind the scenes.


## Configuring a MoneroNode

The `MoneroNode` definition supports the following fields:

  - [`apiVersion`][kubernetes-overview] - Specifies the API version, for example
    `tekton.dev/v1beta1`.
  - [`kind`][kubernetes-overview] - Identifies this resource object as a `MoneroNode` object.
  - [`metadata`][kubernetes-overview] - Specifies metadata that uniquely identifies the
    `MoneroNode` object. For example, a `name`.
  - [`spec`][kubernetes-overview] - Specifies the configuration information for
    this `MoneroNode` object. This must include:
    - [`monerod`](#configuring-monerod) - Specifies the configuration for the
      monero daemon and details like related proxies for non-clearnet usage.

[kubernetes-overview]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields
