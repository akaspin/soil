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
meta {
  "groups" = "first,second,third"
  "rack" = "left"
}

resource "range" "port" {
  min = 20000
  max = 23000
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

`resource` `(map: {})`
: Resource definitions.

`system` `(map: {"pod_exec": "ExecStart=/usr/bin/sleep inf"})` 
: System properties. By default only [Pod unit]({{site.baseurl}}/pod/internals) "Exec" is defined.

`pod` `(map: {})`
: Each [pod stansa]({{site.baseurl}}/pod) defines pod in private namespace.