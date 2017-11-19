---
title: Clustering
layout: default
weight: 100
---

# Clustering

```hcl
cluster {
  node_id = "node-1"
  advertise = "127.0.0.1:7654"
  backend = "consul://127.0.0.1:8500/soil"
  ttl = "3m"
  retry = "30s"
}
``` 

Clustering configuration can be changed during Agent run. Agent doesn't require reachable backend before start or reload reconfiguration. Agent will connect then backend is available.

`node_id` `(string: "")`
: Agent ID. Must be unique within cluster.

`advertise` `(string: "localhost:7654")`
: AAdvertised address.

`backend` `(string: "local://localhost/soil")`
: Backend URL in form `type://address[:port]/chroot`. For now only `"consul"` backend type is supported. To disable clustering use `"local"`.

`ttl` `(duration: "3m")`
: TTL for volatile Agent data.

`retry` `(duration: "30s")`
: Time to wait before try to reconnect to backend.
