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

`runtime` (`true`) Defines where pod units will be deployed: in 
runtime `/run/systemd/system` or local `/etc/systemd/system`. This setting also 
tells where to activate each unit in pod.
 
`target` (`multi-user.target`) [Pod unit](/soil/pod/internals) target.

`constraint` Defines pod constraints.
 
`unit` Units definitions.

`blob` Additional files.

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
  create = "start"
  update = "restart"
  destroy = "stop"
  permanent = false
}
```

`source` SystemD unit source. Can be [interpolated]({{site.baseurl}}/pod/interpolation). Soil agent will write interpolated source to disk only if it differs from existent.

`create` (`start`) Systemd command to execute on unit creation.
 
`update` (`restart`) Systemd command to execute on unit update. This command will be triggered only if unit is exists before pod update and interpolated source from pending pod manifest is differs from existent.  
 
`destroy` (`stop`) Systemd command to execute on unit destroy.
 
Available commands for `create`, `update` and `destroy` are: `start`, `stop`, `restart`, `reload`, `try-restart`, `reload-or-restart`, `reload-or-try-restart`.

`permanent` (`false`) Soil agent will enable unit in SystemD. Enabling this setting assumes what `[Install]` section is present in unit source.

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

`source` BLOB source. Can be [interpolated]({{site.baseurl}}/pod/interpolation).

`permissions` (`0644`) BLOB permissions.

`leave` (`false`) Leave BLOB on disk after destroy.
