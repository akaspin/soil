---
title: Resources
layout: default
weight: 20
---

# Resources

All available resources must be configured:

```hcl
resource "range" "ports1" {
  min = 10000
  max = 20000
}
``` 

Resource definition should be `"<nature>" "<kind>"`. Kind must be unique within Agent config. Internal config depends on resource nature.

## Range

`range` resource provides pool of unique positive integers. Ports for example.

```hcl
resource "range" "port" {
  min = 20000
  max = 30000
}

pod "example" {
  resource "port" "80" {}
  unit "example.service" {
    source = <<EOF
    [Service]
    ExecStart=/usr/bin/docker run --rm --name=%p \
      -p ${resource.port.example.80}:80 alpine httpd -f 
    EOF
  }
}
```

### Configuration

`min` `(uint32: 0)` 
: Minimum value in range.

`max` `(uint32: 4294967295)` 
: Minimum value in range.
 
### Values

`allocated` `(true|false)`
: Allocation status.

`value`
: Allocated value.

`failure`
: Error message if allocation failed.
