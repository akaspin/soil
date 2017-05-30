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

## Docker

```
$ sudo docker run -ti --rm --name soil \
    -v /etc/systemd/system:/etc/systemd/system \
    -v /run/systemd/system:/run/systemd/system \
    -v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
    -v /etc/soil:/etc/soil:ro \
    akaspin/soil
```
