---
title: Agent
layout: default
weight: 100
---

# Agent Operations API

`/agent/` API operates lifecycle of specific Soil Agent 

## Reload

|Method |Path|Result
|-
|`GET` |`/v1/agent/reload`|application/json

Equivalent to `SIGHUP` signal. Reloads Agent.

## Stop

|Method |Path|Result
|-
|`GET` |`/v1/agent/stop`|application/json

Equivalent to `SIGTERM` signal. Stops Agent.

## Drain

|Method |Path|Result
|-
|`PUT` |`/v1/agent/drain`|application/json
|`DELETE` |`/v1/agent/drain`|application/json

`PUT` and `DELETE` methods manages Agent drain state. In drain state Agent removes all pods from SystemD.
