---
title: Pod
layout: default

weight: 200
---

# Pod

Soil uses pods manifests to operate with systemd units and files.

```hcl
pod "my-pod" {
  runtime = true
  target = "default.target"
  constraint {
    "my" = "~ ${meta.groups}"
  }
  
  unit "my-unit-1.service" {
    permanent = false
    create = "start"
    update = "restart"
    destroy = "stop"
    source = <<EOF
      [Unit]
      Description=%p with ${blob.etc-my-pod-sample}
      
      [Service]
      ExecStartPre=/usr/bin/cat /etc/my-pod/sample
      ExecStart=/usr/bin/sleep inf
    EOF
  }
  
  blob "/etc/my-pod/sample" {
    leave = false
    permissions = 0644
    source = <<EOF
      hello
    EOF
  }
}
```

All properties is optional. In fact you can define empty pod without anything.

`runtime` `(bool: true)` 
: Defines where pod units will be deployed: in runtime `/run/systemd/system` or local `/etc/systemd/system`. This setting also tells where to activate each unit in pod.

`target` `(string: "multi-user.target")` 
: [Pod unit]({{site.baseurl}}/pod/internals) target.

`constraint` `(map: {})`
: Defines pod deployments [constraints]({{site.baseurl}}/pod/constraint).

`unit` `(map: {})` 
: Units definitions.

`blob` `(map: {})`
: File definitions.

## Units

All units in pod are defined by `pod` stansa. Units can be added or removed in existent pod on update. 

```hcl
unit "my-unit-1.service" {
  source = <<EOF
    [Unit]
    Description=%p with ${blob.etc-my-pod-sample}
      
    [Service]
    ExecStartPre=/usr/bin/cat /etc/my-pod/sample
    ExecStart=/usr/bin/sleep inf
  EOF
  permanent = false
  create = "start"
  update = "restart"
  destroy = "stop"
}
```

`source` `(string: "")` 
: SystemD unit source. Can be [interpolated]({{site.baseurl}}/pod/interpolation). Soil agent will write interpolated source to disk only if it differs from existent.

`permanent` `(bool: false)` 
: Soil agent will enable unit in SystemD. Enabling this setting assumes what `[Install]` section is present in unit source.

`create` `(string: "start")` 
: Systemd command to execute on unit creation.

`update` `(string: "restart")` 
: Systemd command to execute on unit update. This command will be triggered only if unit is exists before pod update and interpolated source from pending pod manifest is differs from existent.
  
`destroy` `(string: "stop")`
: Systemd command to execute on unit destroy.
 
Available commands for `create`, `update` and `destroy` are: `start`, `stop`, `restart`, `reload`, `try-restart`, `reload-or-restart`, `reload-or-try-restart`. Use empty value `("")` to disable command execution.


## BLOBs

Soil Agent can interact with additional blobs defined by `blob` stansa. BLOBs can be added or removed in existent pod on update.

```hcl
blob "/etc/my-pod/sample" {
  source = <<EOF
    hello
  EOF
  leave = false
  permissions = 0644
}
```

`source` `(string: "")`
: BLOB source. Can be [interpolated]({{site.baseurl}}/pod/interpolation).

`permissions` `(int: 0644)`
: BLOB permissions. All files are deployed with Soil process owner.

`leave` `(bool: false)` 
: Leave BLOB on disk after destroy.

## Mark

Each pod has calculated mark which depends on pod definition.