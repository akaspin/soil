---
title: Registry
layout: default
weight: 200
---

# Registry API

`/registry` API operates with Agent registry.

## Retrieve Pods Manifests

|Method |Path|Result
|-
|`GET` |`/v1/registry`|application/json

Retrieves pods manifests from registry.

### Sample Request

```shell
$ curl -XPUT -d @sample.json http://127.0.0.1:7654/v1/registry
```

### Sample Response

```json
{
  "private": [
    {
      "Namespace": "private",
      "Name": "private-1",
      "Runtime": false,
      "Target": "multi-user.target",
      "Constraint": {
        "${meta.test}": "a"
      },
      "Units": [
        {
          "Create": "start",
          "Update": "restart",
          "Destroy": "stop",
          "Permanent": true,
          "Name": "private-1-1.service",
          "Source": "[Unit]\nDescription=%p\n\n[Service]\n# ${NONEXISTENT}\nExecStart=/usr/bin/sleep inf\n\n[Install]\nWantedBy=multi-user.target\n"
        }
      ],
      "Blobs": null
    }
  ],
  "public": [
    {
      "Namespace": "public",
      "Name": "public-1",
      "Runtime": true,
      "Target": "multi-user.target",
      "Constraint": {
        "${meta.test}": "a"
      },
      "Units": [
        {
          "Create": "start",
          "Update": "restart",
          "Destroy": "stop",
          "Permanent": false,
          "Name": "public-1-1.service",
          "Source": "[Unit]\nDescription=%p\n\n[Service]\n# ${NONEXISTENT}\nExecStart=/usr/bin/sleep inf\n\n[Install]\nWantedBy=multi-user.target\n"
        }
      ],
      "Blobs": null
    }
  ]
}
```

### Submit Pods Manifests

|Method |Path|Result
|-
|`PUT` |`/v1/registry`|application/json

Submit pod manifests to public namespace.

### Sample Request

```shell
$ curl -XPUT -d @sample.json http://127.0.0.1:7654/v1/registry
```

### Sample Payload

```json
[
  {
    "Namespace": "public",
    "Name": "public-1",
    "Runtime": true,
    "Target": "multi-user.target",
    "Constraint": {
      "${meta.test}": "a"
    },
    "Units": [
      {
        "Create": "start",
        "Update": "restart",
        "Destroy": "stop",
        "Permanent": false,
        "Name": "public-1-1.service",
        "Source": "[Unit]\nDescription=%p\n\n[Service]\n# ${NONEXISTENT}\nExecStart=/usr/bin/sleep inf\n\n[Install]\nWantedBy=multi-user.target\n"
      }
    ],
    "Blobs": null
  }
]
```


## Delete Pods Manifests

|Method |Path|Result
|-
|`DELETE` |`/v1/registry`|application/json

Delete pod manifests from public namespace.

### Sample Request

```shell
$ curl -XDELETE -d `["one","two"]` http://127.0.0.1:7654/v1/registry
```

