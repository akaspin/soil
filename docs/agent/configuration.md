---
title: Configuration
layout: default
weight: 10
---

# Agent configuration

Soil uses command line options and configuration files. Configuration from 
files can be reloaded on `SIGHUP`.

```shell
$ soil agent --id=agent-1 --config=... --meta=rack=left
```

## Command line options

`id` (`localhost`) Agent ID unique within cluster.

`config` (`etc/soil/config.hcl`) Path to agent configuration file. This option can be repeated many times. Agent will parse configuration files in defined order. Each configuration file will be merged with previous.

`meta` Initial `meta` values. This option can be repeated many times.

`address` (`:7654`) Address to listen for API calls.  


## Configuration files

Soil accepts configurations in HCL and JSON.

```hcl
meta {
  "groups" = "first,second,third"
  "rack" = "left"
}

drain = false
exec = "ExecStart=/usr/bin/sleep inf"

pod "first-pod" {
  // ...
}

pod "second-pod" {
  // ...
}
```

`meta` Agent metadata. These values can be used in pod [constraints]({{site.baseurl}}/pod/constraint) and [interpolations]({{site.baseurl}}/pod/interpolation) as `${meta.<key>}`.

`exec` (`ExecStart=/usr/bin/sleep inf`) [Pod unit]({{site.baseurl}}/pod/internals) "Exec" lines.

`pod` Each [pod stansa]({{site.baseurl}}/pod) defines pod in private namespace.