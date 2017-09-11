## 0.2.4

### API

* /v1/agent/drain (PUT, DELETE)
* /v1/status/drain (GET)

### Agent

* status -> allocation
* status.<pod> = present -> allocation.<prod>.present = true

### Clustering

* Initial implementation

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