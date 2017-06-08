---
title: Consul cluster
layout: default
weight: 0
---

# Consul cluster with rolling updates

Consul is great tool. But as raft cluster with strong consistency it requires very complex management. In this recipe we'll deploy Consul cluster with rolling upgrades for server nodes.

## Weapons

1. Dedicated Consul cluster for distributed locks.
1. `smlr` for wait Consul leadership.
1. `systemd-command` for execute commands under Consul lock.

## Lock cluster

Lock cluster is very simple. It doesn't hold any valuable data except locks. We'll use official Consul Docker image. But because lock cluster will be deployed on same instances we'll use different ports.

```hcl
meta {
  "consul_lock" = "true" 
  "consul_lock_server" = "true" 
}

pod "consul-lock-server" {
  blob "/var/lib/soil/consul-lock-server.json" {
    source = <<EOF
    {
      "data_dir": "/var/lib/consul",
      "log_level": "ERROR",
      "ui": false,
      "disable_remote_exec": true,
    
      "node_name": "${agent.id}",
      "retry_join": [],
      "bind_addr": "{{ env "CONSUL_BIND_IPV4" }}",
      "client_addr": "0.0.0.0",
      "ports": {
          "serf_lan": {{ env "CONSUL_SERF_LAN_PORT" }},
          "serf_wan": {{ env "CONSUL_SERF_WAN_PORT" }},
          "server": {{ env "CONSUL_RPC_PORT" }},
          "rpc": {{ env "CONSUL_CLIENT_RPC_PORT" }},
          "http": {{ env "CONSUL_HTTP_PORT" }},
          "dns": {{ env "CONSUL_DNS_PORT" }}
      }
}
    EOF
  }  
}
```