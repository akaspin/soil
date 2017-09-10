---
title: Agent
layout: default
weight: 100
---

# Agent Operations API

`/agent/` API operates lifecycle of specific Soil Agent 

## Reload

#### `GET /v1/agent/reload`

Equivalent to `SIGHUP` signal. Reloads Agent.

## Stop

#### `GET /v1/agent/stop`

Equivalent to `SIGTERM` signal. Stops Agent.
