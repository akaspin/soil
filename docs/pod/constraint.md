# Pod constraints

Each pod can be constrained to agent metadata like:

```hcl
constraint {
  "${meta.machine}" = "big-server"
  "my-group-1,my-group-2" = "~ ${meta.groups}"
  "${meta.index}" = "< 2"
}
```

Soil will deploy pod only if all constraints are met. Both left and right 
values of constraint can me [interpolated](/soil/pod/interpolation).

Also right value of constraint can be prefixed with operation:

`<` or `>` Soil tries to convert values to number and compare them. If at least 
one of values can't be converted Soil compares values as strings in 
lexicographical order. 

`~` Subset operation. This constraint assumes what all values from left subset 
are present in right subset. Subsets are delimited by comma.