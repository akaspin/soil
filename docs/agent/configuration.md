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

## Configuration files

Soil accepts configurations in HCL and JSON.

```hcl
system {
  pod_exec = "ExecStart=/usr/bin/sleep inf"
}

cluster {
  node_id = "node-1"
  advertise = "127.0.0.1:7654"
  backend = "consul://127.0.0.1:8500/soil"
  ttl = "3m"
  retry = "30s"
}

meta {
  "groups" = "first,second,third"
  "rack" = "left"
}

resource "range" "port" {
  min = 20000
  max = 23000
}


pod "first-pod" {
  // ...
}

pod "second-pod" {
  // ...
}
```

`system` `(map: {"pod_exec": "ExecStart=/usr/bin/sleep inf"})` 
: System properties. By default only [Pod unit]({{site.baseurl}}/pod/internals) "Exec" is defined.

`cluster`
: [Clustering]({{site.baseurl}}/agent/clustering) configuration

`meta` `(map: {})` 
: Agent metadata. These values can be used in pod [constraints]({{site.baseurl}}/pod/constraint) and [interpolations]({{site.baseurl}}/pod/interpolation) as `${meta.<key>}`.

`resource` 
: [Resources]({{site.baseurl}}/agent/resources) configurations.

`pod`
: Each [pod stansa]({{site.baseurl}}/pod) defines pod in private namespace.
