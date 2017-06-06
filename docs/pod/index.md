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
 
`target` (`default.target`) [Pod unit](/soil/pod/internals) target.

`constraint` Defines pod constraints.
 
`unit` Units definitions.

`blob` Additional files.