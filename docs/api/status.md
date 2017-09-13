---
title: Status
layout: default
weight: 0
---

# Agent Status API

Status API retrieves status of specific Soil Agent.

## Healthcheck

#### `GET /v1/status/ping`

Returns `200/OK` if agent is alive.

## Drain

#### `GET /v1/status/drain`

Returns `200/OK` and Agent Drain state:

```json
{
  "AgentId": "node-1",
  "Drain": false
}
```