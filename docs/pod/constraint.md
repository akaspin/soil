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

## Operations

```hcl
"value" = "value"   // ok
"value" = "other"   // fail
```

By default soil compares right and left operands for equality. This behaviour can be changed by operation in right value delimited by space.  

```hcl
"value" = "= other"   // fail
"value" = "!= other"  // ok
```

Equal, not equal (`=`, `!=`) Checks values for equality or not equality respectively.
 
```hcl
"0" = "< 2"     // ok
"0" = "<= 2"    // ok
"0" = "> 2"     // fail
"0" = ">= 2"    // fail
```

Less, less or equal, greater, greater or equal (`<`, `>`, `>=`, `<=`) Soil tries to convert values to number and compare them. If at least one of values can't be converted Soil compares values as strings in lexicographical order. 

```hcl
"one,two" = "~ one,two,three"   // ok
"one,two" = "~ two,three"       // fail
```

In (`~`) This constraint assumes what all values from left subset are present in right subset. Subsets are delimited by comma.

```hcl
"one,two" = "!~ one,two,three"   // fail
"one,two" = "!~ two,three"       // fail
"one,two" = "!~ three,four"      // ok
```

Not in `!~` This constraint assumes what none of values from left subset are present in right subset. Subsets are delimited by comma.

## Default constraints

Default constraints are defined for each pod and cannot be changed.

`"${agent.drain}" = "false"` All pods managed by Agent in `drain` state will be destroyed.