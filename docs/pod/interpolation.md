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

## Interpolated Areas

* Constraint fields. Both left and right
* `unit` and `blob` names.
* `unit` and `blob` sources.

## `meta`

`meta` variables can be declared in [Agent configuration]({{site.baseurl}}/agent/configuration). Can be referenced in `constraint`, `unit->source` and `blob->source` areas.

## `pod`

`pod` variables depends on Pod properties. All pod variables are *not* accessible in `constraint` area.

|Variable   |Areas
|-
|`name`, `namespace`  | `unit->{source,name}`, `blob->{source,name}`
|`target`| `unit->source`, `blob->source`

## `blob`

If pod contains one or more BLOBs their hashes will be available as `${blob.<blob-id>}`. There `blob-id` is escaped path. For example blob with path `/etc/my/blob.env` hash will be available in units as `${blob.etc-my-blob.env}`. `blob` variables can be referenced only in `unit->source`.

## `resource`

Any 

## `agent`

Agent variables are accessible as `${agent.*}`.

|Variable   |Description
|-
|`id`| Agent ID
|`advertise`| Advertise address
|`version`| Soil Agent version
|`api`|API Revision
|`drain`|Drain mode.

All `agent` variables can be referenced in in `constraint`, `unit->source` and `blob->source` areas


## `system`

|Variable   |Description
|-
|`pod_exec`| Pod unit "Exec*"

All `system` variables can be referenced in in `constraint`, `unit->source` and `blob->source` areas
