---
title: Constraints
layout: default
weight: 10
---

# Pod constraints

Each pod can be constrained to agent metadata like:

```hcl
constraint {
  "${meta.machine}" = "big-server"
  "my-group-1,my-group-2" = "~ ${meta.groups}"
  "${meta.index}" = "< 2"
}
```

Soil will deploy pod only if all constraints are met. Both left and right values of constraint can me [interpolated](/soil/pod/interpolation). If pod is already deployed but constraints become fail it will be destroyed and vise versa

Also right value of constraint can be prefixed with operation:

`<` or `>` Soil tries to convert values to number and compare them. If at least one of values can't be converted Soil compares values as strings in lexicographical order. 

`~` Subset operation. This constraint assumes what all values from left subset are present in right subset. Subsets are delimited by comma.

`!~` Not subset operation. This constraint assumes what none of values from left subset are present in right subset. Subsets are delimited by comma.

## Default constraints

Default constraints are defined for each pod:

`"${agent.drain}" = "false"` All pods managed by Agent in `drain` state will be destroyed.