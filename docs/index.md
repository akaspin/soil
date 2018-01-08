---
title: Soil
layout: landing
index: true
weight: 0
---

# Soil

#### Behavioral SystemD provision

---

## Schema-based

Soil maintains the specified schema. Just change schema and Soil will sync everything.
Define metadata for constraints and interpolation. All provision will be changed in dynamic way.

---

## Behavioral 

| Soil never touch unchanged entities. All defined SystemD units have a specified behavior depending on the deployment: create, update, destroy. | ![Pod lifecycle]({{site.baseurl}}/assets/images/noun_1437720_cc.svg){: .test-image } |

## Run anywhere

Soil has only one dependency - SystemD socket. It can be deployed in unprivileged Docker container.

## Node-centric

Soil doesn't need masters to run. Clustering is optional. Need clustering - deploy any KV store by Soil.

 
