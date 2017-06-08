---
title: Internals
layout: default
weight: 100
---

# Pod internals

Soil always deploy one additional unit for each pod. Soil uses this unit to hold pod metadata and recover state after agent restarts.

```
### POD my-pod {"AgentMark":...,"Namespace":"private","PodMark":...}
### UNIT /run/systemd/system/my-unit.service {"Create":"start","Update":"restart","Destroy":"stop","Permanent":false}
### BLOB /etc/my-pod/sample {"Leave":false,"Permissions":420}
[Unit]
Description=my-pod
Before=my-unit.service 
[Service]
ExecStart=/usr/bin/sleep inf
[Install]
WantedBy=multi-user.target
```

Name of this unit is depends on unit name and namespace like `pod-private-my-pod.service`.
 
`ExecStart` lines can be configured by [`exec`]({{site.baseurl}}/agent/configuration) agent configuration setting.