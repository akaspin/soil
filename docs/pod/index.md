---
title: pod manifest
---

# Pod manifest

```hcl
pod "my-pod" {
  runtime = true
  target = "default.target"

  constraint {
    "${meta.my-pod}" = "true"
    "${agent.id}" = "localhost"
  }

  unit "my-1.service" {
    permanent = false
  
    create = ""
    update = ""
    destroy = "stop"
    
    source = <<EOF
      [Unit]
      Description=%p
      
      [Service]
      ExecStart=/usr/bin/sleep inf
      
      [Install]
      WantedBy=default.target
    EOF
  }
}

```
