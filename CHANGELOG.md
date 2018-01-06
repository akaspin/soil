## 0.5.1 (06.01.2018)

### Features

* Handle invalid resource providers

## 0.5.0 (28.12.2017)

### Breaking changes

* Resource provision is fully refactored. From this release all resources use
  providers as source. Providers are located inside pods and obey pod 
  constraints.

## 0.4.2 (24.11.2017)

### Features

* `provision` variables for interpolations
* default values in interpolations
* `/v1/registry` API

## 0.4.1 (19.11.2017)

This release reintroduces clustering.

### Features

* Clustering
* (API) `GET` `/v1/cluster/nodes`

### Deprecated

* `--id` command-line option for `soil agent` 

## 0.4.0 (26.10.2017)

This release introduces resources.

### Breaking changes

* `/v1/status/ping`, `/v1/agent/reload` and `/v1/agent/drain` endpoints are 
  available during refactoring.
* Clustering is temporary disabled.

### Features

* Resources

## 0.3.2 (06.10.2017)

### Features

* Allow to use `pod.{name,namespace}` in `unit` and `blob` names.
* Allow to use `pod.{name,namespace,target}` in `unit` and `blob` sources.

## 0.3.1 (27.09.2017)

### Breaking changes

* (API) GET /v1/agent/reload -> PUT /v1/agent/reload
* (API) GET /v1/agent/stop -> PUT /v1/agent/stop

### New features

* (API) {GET|PUT|DELETE} /v1/registry/pods

## 0.3.0 (23.09.2017)

### API

* /v1/agent/drain (PUT, DELETE)
* /v1/status/node (GET)
* /v1/status/nodes (GET)

### Agent

* status -> allocation
* status.<pod> = present -> allocation.<prod>.present = true
* Massive refactorings

### Clustering

* Announce node
* Sync announced nodes

## 0.2.3 (13.09.2017)

### Agent

* Use strict drain state
* Remove loud fields from pod status 

## 0.2.2 (11.09.2017)

### Test

* Refactor test system

## 0.2.1 (10.09.2017)

### API

* /v1/status/ping
* /v1/agent/reload
* /v1/agent/stop

## 0.2.0 (10.09.2017)

* Evaluator runs instructions simultaneously in sequential phases
* Temporary remove Worker pools

## 0.1.0

Initial release
