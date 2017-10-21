# Resource internals

## Provision manager resource fields

1. `[resource].<kind>.<pod-name>.<resource-name>.allocated`: `true`
1. `[resource].<kind>.<pod-name>.<resource-name>.*`: values
1. `[__resource.values].<kind>.<pod-name>.<resource-name>` Resource values in JSON
1. `[__resource.state].clean`: `true` global

## Resource manager

1. `[__resource.kind].<kind>.allow`: `true`
1. `[__resource.request].allow`: `true`