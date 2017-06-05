# Pod

Soil uses pods manifests.

```hcl
pod "my-pod" {
  
  // 
  runtime = true
  target = "default.target"
  constraint {
    "my" = "~ ${meta.groups}"
  }
  
  unit "my-unit.service" {
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

