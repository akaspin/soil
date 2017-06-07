---
title: Agent
layout: default
weight: 100
---


# Agent

Soil Agent ships as single static-linked binary. It interacts with DBus manages systemd units in `system` runtime and local directories. Also Agent can operate with additional BLOBs defined in [pod manifests]({{site.baseurl}}/pod)

## Required paths

Soil Agent needs RW access to next system paths:

`/var/run/dbus/system_bus_socket` Path to DBus socket.

`/run/systemd/system` Systemd Runtime path.

`/etc/systemd/system` Systemd Local path

Also Soil Agent needs RO access to own configuration files and RW access to directories where agent manages BLOBs.
