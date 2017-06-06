---
title: Lifecycle
layout: default
weight: 30
---

# Pod lifecycle

> If pod update changes `pod->runtime` all pod will be destroyed and recreated in corresponding SystemD space.

Soil Agent evaluates each pod change in next order:

1. Executes systemd commands defined in `unit->destroy` for units which be destroyed.
2. Disables and destroys (deletes from filesystem) all units which be destroyed.
3. Deletes all BLOBs which be destroyed.
4. Stores all BLOBs which be created or updated.
5. Writes and optionally enables all units which be created or updated.
6. Executes systemd corresponding commands from `unit->create|update` for all units which be created or updated. 

![Pod lifecycle]({{site.baseurl}}/assets/images/pod-evaluation.svg)

Stages are optional. Pod creation will fire stages `4`, `5` and `6` with commands from `unit->create`. Pod destroy will fire only stages `1`, `2` and `3`. On pod update Soil Agent will calculate plan based on diff from deployed and pending pods.

Note in example above what `unit-4` was not changed on pod update.