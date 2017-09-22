# APIv1

# `/agent`

Single node operations.

```
PUT /drain *
DEL /drain *
?   /reload
?   /stop
```

# `/status`

```
GET /ping
GET /node   - local agent info: attributes, meta, allocations
GET /nodes  - discovered nodes
GET /registry   - pods
```

# `/registry`

```
PUT /pod
DEL /pod
```
