---
title: Consul cluster
layout: default
weight: 0
---

# Consul cluster with rolling updates

Consul is great tool. But as raft cluster with strong consistency it requires very complex management. In this recipe we'll deploy Consul cluster with rolling upgrades for server nodes.

## Weapons

1. Dedicated Consul cluster for distributed locks.
1. Separate server and client Consul nodes.
1. `smlr` for wait Consul leadership.
1. `systemd-command` for execute commands under Consul lock.

