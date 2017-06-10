---
title: Interpolation
layout: default

weight: 20
---

# Interpolation

Soil agent interpolates variables in pod constraints, `unit->source` and `blob->source` declared as `${source.variable}`. If variable is not defined Soil agent leaves it unchanged.  

```hcl
pod "my-pod" {
  constraint {
    "${meta.rack}" = "rack-1"
  }

  unit "unit-1" {
    source = <<EOF
    # ${meta.rack}
    [Service]
    EnvironmentFile=/etc/test
    ...
    EOF
  }
  blob "/etc/test" {
    source = <<EOF
    RACK=${meta.rack}
    EOF
  }
}
```

All interpolated variables are named as `<source-name>.<variable-name>`.

## `meta` source

`meta` variables can be declared in [Agent configuration]({{site.baseurl}}/agent/configuration).

## `blob` source

If pod contains one or more BLOBs their hashes will be available as `${blob.<blob-id>}`. There `blob-id` is escaped path. For example blob with path `/etc/my/blob.env` hash will be available in units as `${blob.etc-my-blob.env}`.

## `agent` source

Agent variables are accessible as `${agent.*}`:

`id` Agent ID.

`drain` Agent drain state.

`pod_exec` Pod unit "Exec*".

## `status` source

All deployed pods are registered in `status` source. All values for pod are prefixed with pod name.

`<pod-name>` Pod presence. Can be only `present`. If pod is not deployed value will be absent.
 
`<pod-name>.units` List of pod [units]({{site.baseurl}}/pod/#units) delimited by comma.

`<pod-name>.namespace` Pod [namespace]({{site.baseurl}}/agent/namespaces).

`<pod-name>.mark` Pod [mark]({{site.baseurl}}/pod/#mark).

`<pod-name>.agent_mark` Agent [mark]({{site.baseurl}}/agent/#mark).

`<pod-name>.failures` Failures what happened on pod deployment.

