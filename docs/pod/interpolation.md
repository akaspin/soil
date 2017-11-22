---
title: Interpolation
layout: default

weight: 20
---

# Interpolation

Soil Agent interpolates variables in pod manifests. Variables can be referenced by `${source.any.variable}`. If variable is not defined Soil agent leaves it unchanged.  

```hcl
pod "my-pod" {
  constraint {
    "${meta.rack}" = "rack-1"
    "${provision.other-pod.present}" = "true"
    "${provision.other-pod.state}" = "!= destroy"
    "${meta.with.default|yes}" = "yes"
  }

  unit "${pod.name}-unit-1" {
    source = <<EOF
    # ${meta.rack}
    [Service]
    EnvironmentFile=/etc/test
    ...
    EOF
  }
  blob "/etc/test/${pod.namespace}" {
    source = <<EOF
    RACK=${meta.rack}
    EOF
  }
}
```

Interpolation may be defined with default value. Default value is constant delimited by pipe sign (`|`). If variable is not defined Soil will use default value.

## Interpolated Areas

* Constraint fields. Both left and right
* `unit` and `blob` names.
* `unit` and `blob` sources.

## `meta`

`meta` variables can be declared in [Agent configuration]({{site.baseurl}}/agent/configuration). Can be referenced in `constraint`, `unit->source` and `blob->source` areas.

## `pod`

`pod` variables depends on Pod properties. All `pod` variables are *not* accessible in `constraint` area. `pod` variables are not available for other pods.

|Variable   |Areas
|-
|`name`, `namespace`  | `unit->{source,name}`, `blob->{source,name}`
|`target`| `unit->source`, `blob->source`

## `blob`

If pod contains one or more BLOBs their hashes will be available as `${blob.<blob-id>}`. There `blob-id` is escaped path. For example blob with path `/etc/my/blob.env` hash will be available in units as `${blob.etc-my-blob.env}`. `blob` variables can be referenced only in `unit->source`. `blob` variables are not available for other pods.

## `resource`

All allocated resources can be referenced as `${resource.<kind>.<pod>.<resource>.*}`

|Variable   |Areas
|-
|`allocated`:`{true,false}`    | `constraint`
|`*`                           | `constraint`, `unit->source`, `blob->source`
|`failure`                     | `constraint`, `unit->source`, `blob->source`

All resources are available within all pods on Agent. If resource can't be allocated it will be marked with `allocated`:`false` and `failure` with error.

## `provision`

Scheduler reports about provision states for all pods to `${provision.<pod-name>.*}`. These variables are available in constraints for all pods. 

|Variable   |Description
|-
|`present`                                      |Pod is present in provision scheduler
|`state`:`{done,create,update,destroy,dirty}`   |Provision state 

## `system`

|Variable   |Description
|-
|`pod_exec`| Pod unit "Exec*"

All `system` variables can be referenced in in `constraint`, `unit->source` and `blob->source` areas
