REPO	= github.com/akaspin/soil
BIN		= soil

BENCH	= .
TESTS	= .
TEST_TAGS =
TEST_ARGS =

CWD 		= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VENDOR 		= $(CWD)/vendor
SRC 		= $(shell find . -type f \( -iname '*.go' ! -iname "*_test.go" \) -not -path "./vendor/*")
SRC_TEST 	= $(shell find . -type f -name '*_test.go' -not -path "./vendor/*")
SRC_VENDOR 	= $(shell find ./vendor -type f \( -iname '*.go' ! -iname "*_test.go" \))
PACKAGES    = $(shell cd $(GOPATH)/src/$(REPO) && go list ./... | grep -v /vendor/)

V=$(shell git describe --always --tags --dirty)
GOOPTS=-installsuffix cgo -ldflags '-s -w -X $(REPO)/command.V=$(V)'

GOBIN ?= $(GOPATH)/bin


sources: $(SRC) $(SRC_TEST)
	go vet $(PACKAGES)
	go fmt $(PACKAGES)

###
### Test
###

test: test-unit test-systemd test-integration

###
### Test Unit
###

test-unit: test-unit-simple test-unit-cluster

test-unit-simple: $(SRC) $(SRC_TEST)
	go test -run=$(TESTS) $(TEST_ARGS) -tags="test_unit $(TEST_TAGS)" $(PACKAGES)

test-unit-cluster: $(SRC) $(SRC_TEST)
	go test -run=$(TESTS) $(TEST_ARGS) -tags="test_cluster $(TEST_TAGS)" $(PACKAGES)


clean-test-unit:
	find . -name .consul_data_* -type d -exec rm -rf {} +

###
### Test SystemD
###

test-systemd: testdata/systemd/.vagrant-ok
	docker -H 127.0.0.1:2475 run --rm --name=test \
		-v /run/soil:/run/soil \
		-v /var/lib/soil:/var/lib/soil \
		-v /run/systemd/system:/run/systemd/system \
		-v /etc/systemd/system:/etc/systemd/system \
		-v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
		-v /vagrant:/go/src/github.com/akaspin/soil \
		golang:1.9 go test -run=$(TESTS) -p=1 $(TEST_ARGS) -tags="test_systemd $(TEST_TAGS)" $(PACKAGES)

testdata/systemd/.vagrant-ok: testdata/systemd/Vagrantfile
	cd testdata/systemd && vagrant up --parallel

clean-test-systemd:
	cd testdata/systemd && vagrant destroy -f
	rm -rf testdata/systemd/.vagrant*

###
### Test Integration
###

test-integration: \
	test-integration-env-up-1 \
	test-integration-env-up-2 \
	test-integration-env-up-3
	go test -run=$(TESTS) -p=1 $(TEST_ARGS) -tags="test_integration $(TEST_TAGS)" $(PACKAGES)

test-integration-env-up-%: \
		testdata/integration/.vagrant-ok \
		dist/$(BIN)-$(V)-linux-amd64.tar.gz
	HOST=172.17.8.10$* AGENT_ID=node-$* V=$(V) docker-compose -H 127.0.0.1:257$* -f testdata/integration/compose.yaml up -d --build

testdata/integration/.vagrant-ok: testdata/integration/Vagrantfile
	cd testdata/integration && vagrant up --parallel

integration-env-down:
	docker-compose -H 127.0.0.1:2571 -f testdata/integration/compose.yaml down --rmi all
	docker-compose -H 127.0.0.1:2572 -f testdata/integration/compose.yaml down --rmi all
	docker-compose -H 127.0.0.1:2573 -f testdata/integration/compose.yaml down --rmi all

clean-test-integration:
	cd testdata/integration && vagrant destroy -f
	rm -rf testdata/integration/.vagrant*

###
### Dist
###

check-src: $(SRC) $(SRC_TEST)
	go vet $(PACKAGES)
	[[ -z `gofmt -d -s -e $^` ]]

dist: \
	dist/$(BIN)-$(V)-darwin-amd64.tar.gz \
	dist/$(BIN)-$(V)-linux-amd64.tar.gz

dist/$(BIN)-$(V)-%-amd64.tar.gz: dist/%/$(BIN) dist/%/$(BIN)-debug
	tar -czf $@ -C ${<D} $(notdir $^)

dist/%/$(BIN): $(SRC) $(SRC_VENDOR)
	@mkdir -p $(@D)
	GOPATH=$(GOPATH) CGO_ENABLED=0 GOOS=$* go build $(GOOPTS) -o $@ $(REPO)/command/$(BIN)

dist/%/$(BIN)-debug: $(SRC) $(SRC_VENDOR)
	@mkdir -p $(@D)
	GOPATH=$(GOPATH) CGO_ENABLED=0 GOOS=$* go build $(GOOPTS) -tags debug -o $@ $(REPO)/command/$(BIN)

docker-image: dist/$(BIN)-$(V)-linux-amd64.tar.gz
	docker build --build-arg V=$(V) -t soil-local:$(V) -f Dockerfile.local .

clean-dist:
	rm -rf dist

###
###	Install
###

install: $(GOBIN)/$(BIN)
install-debug: $(GOBIN)/$(BIN)-debug

$(GOBIN)/$(BIN): $(SRC)
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build $(GOOPTS) -o $@ $(REPO)/command/$(BIN)

$(GOBIN)/$(BIN)-debug: $(SRC)
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build $(GOOPTS) -tags debug -o $@ $(REPO)/command/$(BIN)

uninstall:
	rm -rf $(GOBIN)/$(BIN)
	rm -rf $(GOBIN)/$(BIN)-debug

###
### clean
###

clean: clean-dist uninstall clean-test-unit clean-test-systemd clean-test-integration clean-docs

###
### docs
###

docs:
	docker run --rm -v $(CWD)/docs:/site -p 4000:4000 andredumas/github-pages serve --watch

clean-docs:
	rm -rf docs/_site


.PHONY: \
	docs \
	test \
	test-unit \
	test-systemd \
	test-integration \
	clean