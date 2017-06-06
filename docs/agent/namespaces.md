---
title: Namespaces
layout: default
weight: 200
---

# Agent namespaces

> Only "private" is supported at this time.

Agent registers pods in two namespaces: "private" and "public".

Pods in "private" namespace are defined in agent configuration files. Soil 
agent begins manage pods in "private" namespace after start regardless of 
cluster state. Pods in "private" namespace can't use counter constraints and 
not replicated in cluster.

"public" namespace is set of pods which submitted externally. "public" 
namespace is replicating between all agents in cluster. Also pods in "public"
namespace can use counter constraints.

If two pods with one name are defined in both namespaces Soil always prefers 
pod in "private" namespace.
