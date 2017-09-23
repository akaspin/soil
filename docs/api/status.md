---
title: Status
layout: default
weight: 0
---

# Agent Status API

Status API retrieves status of specific Soil Agent.

## Healthcheck

|Method |Path|Result
|-
|`GET` |`/v1/status/ping`|application/json


Returns `200/OK` if agent is alive.

## Node

|Method |Path|Result
|-
|`GET` |`/v1/status/node`|application/json

Returns status of specific Agent:

```json
{
  "agent": {
    "advertise": "127.0.0.1:7654",
    "api": "v1",
    "drain": "false",
    "id": "node-1.node.dc1.consul",
    "pod_exec": "ExecStart=/usr/bin/sleep inf",
    "version": "0.3.1"
  },
  "meta": {
    "rack": "left"
  }
}
```

## Nodes

|Method |Path|Result
|-
|`GET` |`/v1/status/nodes`|application/json

Returns properties of all discovered nodes:

```json
[
  {
    "Id": "node-3.node.dc1.consul",
    "Advertise": "127.0.0.1:7654",
    "Drain": "false",
    "Version": "0.2.3-17-g0031ee6-dirty",
    "API": "v1"
  },
  {
    "Id": "node-1.node.dc1.consul",
    "Advertise": "127.0.0.1:7654",
    "Drain": "false",
    "Version": "0.2.3-17-g0031ee6-dirty",
    "API": "v1"
  },
  {
    "Id": "node-2.node.dc1.consul",
    "Advertise": "127.0.0.1:7654",
    "Drain": "false",
    "Version": "0.2.3-17-g0031ee6-dirty",
    "API": "v1"
  }
]
```