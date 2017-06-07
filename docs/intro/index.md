---
title: Intro
layout: default
weight: 0
---

# Introduction

Soil is designed to ensure a correct and predictable lifecycle for set of systemd units on completely empty machines. Soil has no dependencies except systemd.

## Pods and unit lifecycle

Soil groups managed units into [pods]({{site.baseurl}}/pod). All units in one pod will be deployed before Soil send command to one of them. 

Pods can be constrained to different aspects such metadata.

Each unit in pod has [lifecycle-based]({{site.baseurl}}/pod/lifecycle) behaviour. Soil can run specific commands on unit create, update and destroy. Instead destroy and recreate pod on update Soil evaluates differences between units and updates only changed.

## Heavy duty

Soil has no dependencies and not requires strong cluster consistency. Each [Soil Agent]({{site.baseurl}}/agent) can be restarted without any consequences. Soil Agent doesn't require access to low level system partitions like `/proc` and can be deployed in unprivileged docker container. 


