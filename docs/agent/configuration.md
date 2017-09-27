---
title: Configuration
layout: default
weight: 10
---

# Agent configuration

Soil uses command line options and configuration files. Configuration from 
files can be reloaded on `SIGHUP`.

```shell
$ soil agent --id agent-1 --config=config.hcl --meta rack=left
```

## Command line options

`id` `(string: "localhost")`
: Agent ID. This value is matters only if Agent uses *public* namespace and should be unique within cluster.

`config` `([]string: "etc/soil/config.hcl")`
: Path to agent configuration file. This option can be repeated many times. Agent will parse configuration files in defined order. Each configuration file will be merged with previous.

`meta` (`[]string: []`)
: Initial values which can be referenced as `${meta.my_value}`. This option can be repeated many times. Definition form is `variable=value`.

`address` (`string: ":7654"`) 
: Address to listen for [API]({{site.baseurl}}/api) calls.  

`public` `(bool: false)`
: Specifies if Agent should use public namespace.

`advertise` `(string: "127.0.0.1:7654")`
: Address to advertise. This address should be reachable from other Agents.

`url` `(string: consul://127.0.0.1:8500/soil)`
: *Public* backend URL in form `type://host[:port]/chroot`. Available backends are `consul`, `etcd` and `zookeeper`. At now only Consul backend is tested. You can define more than one `host[:port]` delimited by comma.

`ttl` `(duration: "3m")`
: TTL for dynamic entities in public backend. Different backends has specific TTL policies.
 
`timeout` `(duration: "1m")`
: Timeout to connect to public backend.

`interval` `duration: "30s"`
: Period to sleep between retries on public backend failures.

## Configuration files

Soil accepts configurations in HCL and JSON.

```hcl
meta {
  "groups" = "first,second,third"
  "rack" = "left"
}

exec = "ExecStart=/usr/bin/sleep inf"

pod "first-pod" {
  // ...
}

pod "second-pod" {
  // ...
}
```

`meta` `(map: {})` 
: Agent metadata. These values can be used in pod [constraints]({{site.baseurl}}/pod/constraint) and [interpolations]({{site.baseurl}}/pod/interpolation) as `${meta.<key>}`.

`system` `(map: {"pod_exec": "ExecStart=/usr/bin/sleep inf"})` 
: System properties. By default only [Pod unit]({{site.baseurl}}/pod/internals) "Exec" is defined.

`pod` `(map: {})`
: Each [pod stansa]({{site.baseurl}}/pod) defines pod in private namespace.