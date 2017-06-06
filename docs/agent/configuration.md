# Agent configuration

Soil uses command line options and configuration files. Configuration from 
files can be reloaded on `SIGHUP`.

```
$ soil agent --id=agent-1 --pool=4 --config=/etc/soil/config.hcl --config=...
```

## Command line options

`id` (`localhost`) Agent ID unique within cluster.

`pool` (`4`) Worker pool size.

`config` (`etc/soil/config.hcl`) Path to agent configuration file. This 
option can be repeated many times. Agent will parse configuration files 
in defined order.

## Configuration files

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
```

