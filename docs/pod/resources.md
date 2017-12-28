---
title: Resources
layout: default
weight: 20
---

# Resources

All available resources must be configured by provider stansa. Once provider is defined pods can request resources. 

```hcl
pod "1" {
  provider "range" "port" {
    min = 10000
    max = 20000
  }
  resource "1.port" "8080" {}
}

pod "2" {
  resource "1.port" "8081" {}
  unit "u1.service" {
    source = <<EOF
      # ${resource.1.8080.value}
      # ${resource.2.8081.value}
    EOF
  }
}
``` 

Providers should be defined as `"kind" "name"`. Resources should reference provider as `<pod>.<provider-name>`. 

## Range

`range` resource provides pool of unique positive integers. Ports for example.

```hcl
pod "example" {
  provider "range" "port" {
    min = 10000
    max = 20000
  }
  resource "example.port" "80" {}
  unit "example.service" {
    source = <<EOF
    [Service]
    ExecStart=/usr/bin/docker run --rm --name=%p \
      -p ${resource.example.port.80}:80 alpine httpd -f 
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

`provider`
: Provider name.

`value`
: Allocated value.

`failure`
: Error message if allocation failed.
