---
title: Deployment
layout: default
weight: 0
---

# Agent deployment

Simplest way to deploy Soil Agent is run it in Docker container:

```
[Unit]
Description=%p

[Service]
SyslogIdentifier=%p
Restart=on-failure
RestartSec=30

ExecStartPre=-/usr/bin/docker pull akaspin/soil:latest
ExecStart=/usr/bin/docker run --rm --name=%p \
    -v /run/systemd/system:/run/systemd/system \
    -v /etc/systemd/system:/etc/systemd/system \
    -v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
    -v /etc/soil:/etc/soil:ro \
    -v /var/lib/soil:/var/lib/soil \
    -v /run/soil:/run/soil \
    akaspin/soil:latest agent \
        --id=%H \
        --config=/etc/soil/base.hcl \
        --config=/etc/soil/field.hcl
ExecReload=/usr/bin/docker kill -s HUP %p

[Install]
WantedBy=multi-user.target
```

This setup will use `/etc/soil` for soil [configuration]({{site.baseurl}}/agent/configuration) files and assumes to deploy BLOBs in `/var/lib/soil` or `/run/soil`.

