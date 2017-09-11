REPO	= github.com/akaspin/soil
BIN		= soil

BENCH	= .
TESTS	= .
TEST_TAGS =

CWD 		= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
VENDOR 		= $(CWD)/vendor
SRC 		= $(shell find . -type f \( -iname '*.go' ! -iname "*_test.go" \) -not -path "./vendor/*")
SRC_TEST 	= $(shell find . -type f -name '*_test.go' -not -path "./vendor/*")
SRC_VENDOR 	= $(shell find ./vendor -type f \( -iname '*.go' ! -iname "*_test.go" \))
PACKAGES    = $(shell cd $(GOPATH)/src/$(REPO) && go list ./... | grep -v /vendor/)

V=$(shell git describe --always --tags --dirty)
GOOPTS=-installsuffix cgo -ldflags '-s -w -X $(REPO)/command.V=$(V)'

GOBIN ?= $(GOPATH)/bin


###
### Test
###

sources: $(SRC) $(SRC_TEST)
	go vet $(PACKAGES)
	go fmt $(PACKAGES)


test: testdata/.vagrant/machines/soil-test/virtualbox/id
	docker -H 127.0.0.1:2475 run --rm --name=test \
		-v /run/soil:/run/soil \
		-v /var/lib/soil:/var/lib/soil \
		-v /run/systemd/system:/run/systemd/system \
		-v /etc/systemd/system:/etc/systemd/system \
		-v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
		-v /vagrant:/go/src/github.com/akaspin/soil \
		golang:1.9 go test -run=$(TESTS) -p=1 -tags="test_unit test_systemd $(TEST_TAGS)" $(PACKAGES)

test-verbose: testdata/.vagrant/machines/soil-test/virtualbox/id
	docker -H 127.0.0.1:2475 run --rm --name=test \
		-v /run/soil:/run/soil \
		-v /var/lib/soil:/var/lib/soil \
		-v /run/systemd/system:/run/systemd/system \
		-v /etc/systemd/system:/etc/systemd/system \
		-v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
		-v /vagrant:/go/src/github.com/akaspin/soil \
		golang:1.9 go test -run=$(TESTS) -p=1 -v -tags="test_unit test_systemd $(TEST_TAGS)" $(PACKAGES)

testdata/.vagrant/machines/soil-test/virtualbox/id: testdata/Vagrantfile
	cd testdata && vagrant up

clean-test:
	cd testdata && vagrant destroy -f
	rm -rf testdata/.vagrant

###
### Integration
###

integration: \
	integration-env-up-1 \
	integration-env-up-2 \
	integration-env-up-3
	go test -run=$(TESTS) -p=1 -v -tags="test_integration $(TEST_TAGS)" $(PACKAGES)

integration-env-up-%: \
		integration/testdata/.vagrant/machines/soil-integration-01/virtualbox/id \
		dist/$(BIN)-$(V)-linux-amd64.tar.gz
	HOST=172.17.8.10$* V=$(V) docker-compose -H 127.0.0.1:257$* -f integration/testdata/compose.yaml up -d --build

integration/testdata/.vagrant/machines/soil-integration-01/virtualbox/id: testdata/Vagrantfile
	cd integration/testdata && vagrant up

integration-env-down:
	docker-compose -H 127.0.0.1:2571 -f integration/testdata/compose.yaml down --rmi all
	docker-compose -H 127.0.0.1:2572 -f integration/testdata/compose.yaml down --rmi all
	docker-compose -H 127.0.0.1:2573 -f integration/testdata/compose.yaml down --rmi all

clean-integration:
	cd integration/testdata && vagrant destroy -f
	rm -rf integration/testdata/.vagrant

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

clean: clean-dist uninstall clean-test clean-integration clean-docs

###
### docs
###

docs:
	docker run --rm -v $(CWD)/docs:/site -p 4000:4000 andredumas/github-pages serve --watch

clean-docs:
	rm -rf docs/_site


.PHONY: \
	docs \
	test test-verbose \
	integration integration-verbose \
	clean