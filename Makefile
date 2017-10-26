REPO	= github.com/akaspin/soil
BIN		= soil

TESTS	      = .
TEST_TAGS     =
TEST_ARGS     =
TEST_PACKAGES ?= $(shell cd $(GOPATH)/src/$(REPO) && go list ./... | grep -v /vendor/)
BENCH	      = .


GO_IMAGE    = golang:1.9.1
CWD 		= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SRC 		= $(shell find . -type f \( -iname '*.go' ! -iname "*_test.go" \) -not -path "./vendor/*")
SRC_TEST 	= $(shell find . -type f -name '*_test.go' -not -path "./vendor/*")
SRC_VENDOR 	= $(shell find ./vendor -type f \( -iname '*.go' ! -iname "*_test.go" \))
PACKAGES    = $(shell cd $(GOPATH)/src/$(REPO) && go list ./... | grep -v /vendor/)

V           = $(shell git describe --always --tags --dirty)
GOOPTS      = -installsuffix cgo -ldflags '-s -w -X $(REPO)/command.V=$(V)'
GOBIN       ?= $(GOPATH)/bin


sources: $(SRC) $(SRC_TEST) 		## go fmt and vet
	go fmt $(PACKAGES)
	go vet $(PACKAGES)

deps: Gopkg.toml					## update vendor
	dep ensure --update -v

###
### Test
###

test: test-unit test-systemd 		## run all tests

clean-test: clean-test-systemd		## clean test artifacts
	find . -name ".test_*" -exec rm -rf {} \;

test-unit: $(SRC) $(SRC_TEST)		## run unit tests
	go test -run=$(TESTS) $(TEST_ARGS) -tags="test_unit $(TEST_TAGS)" $(TEST_PACKAGES)

test-systemd: testdata/systemd/.vagrant-ok	## run SystemD tests
	docker -H 127.0.0.1:2475 run --rm --name=test -v /run/soil:/run/soil -v /var/lib/soil:/var/lib/soil -v /run/systemd/system:/run/systemd/system -v /etc/systemd/system:/etc/systemd/system -v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket -v /vagrant:/go/src/github.com/akaspin/soil $(GO_IMAGE) go test -run=$(TESTS) -p=1 $(TEST_ARGS) -tags="test_systemd $(TEST_TAGS)" $(TEST_PACKAGES)

testdata/systemd/.vagrant-ok: testdata/systemd/Vagrantfile
	cd testdata/systemd && vagrant up --parallel
	touch testdata/systemd/.vagrant-ok

clean-test-systemd:	## clean Systemd tests artifacts
	cd testdata/systemd && vagrant destroy -f
	rm -rf testdata/systemd/.vagrant*

###
### Dist
###

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

clean: clean-dist uninstall clean-docs clean-test

###
### docs
###

docs:
	docker run --rm -v $(CWD)/docs:/site -p 4000:4000 andredumas/github-pages serve --watch

clean-docs:
	rm -rf docs/_site


.PHONY: \
	docs \
	test test-unit \
	test-systemd \
	clean clean-dist uninstall clean-test clean-docs
