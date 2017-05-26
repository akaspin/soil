---
title: soil
---


# Agent deployment

Soil agent shipped as single static binary and require RW access to next resources:

1. System DBus socket (/var/run/dbus/system_bus_socket)
2. Runtime Systemd directory (/run/systemd/system)
3. Local Systemd directory (/etc/systemd/system)

Both Runtime and Local directories are optional.

Also Soil Agent needs RO access to configuration files and private pod manifests. 

## Running

```
$ soil agent --help
Run agent

Usage:
  soil agent [flags]

Flags:
  -c, --config stringArray    configuration file (default [/etc/soil/config.hcl])
  -h, --help                  help for agent
```

`config` flag can be repeated many times. Configuration files are evaluated in 
order. If one or more paths are not exists they will be ignored.

