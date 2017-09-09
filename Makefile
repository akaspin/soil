REPO	= github.com/akaspin/soil
BIN		= soil

BENCH	= .
TESTS	= .
TEST_TAGS = ""

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


test:
	docker -H 127.0.0.1:2375 run --rm \
		-v /run/soil:/run/soil \
		-v /var/lib/soil:/var/lib/soil \
		-v /run/systemd/system:/run/systemd/system \
		-v /etc/systemd/system:/etc/systemd/system \
		-v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
		-v /vagrant:/go/src/github.com/akaspin/soil \
		golang:1.9 go test -run=$(TESTS) -p=1 -tags="$(TEST_TAGS)" $(PACKAGES)

test-verbose:
	docker -H 127.0.0.1:2375 run --rm \
		-v /run/soil:/run/soil \
		-v /var/lib/soil:/var/lib/soil \
		-v /run/systemd/system:/run/systemd/system \
		-v /etc/systemd/system:/etc/systemd/system \
		-v /var/run/dbus/system_bus_socket:/var/run/dbus/system_bus_socket \
		-v /vagrant:/go/src/github.com/akaspin/soil \
		golang:1.9 go test -run=$(TESTS) -p=1 -v -tags="$(TEST_TAGS)" $(PACKAGES)

###
### Dist
###

check-src: $(SRC) $(SRC_TEST)
	go vet $(PACKAGES)
	[[ -z `gofmt -d -s -e $^` ]]

docker-image: dist/$(BIN)-$(V)-linux-amd64.tar.gz
	docker build --build-arg V=$(V) -t soil-local:$(V) -f Dockerfile.local .

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


###
###	Install
###

install: $(GOBIN)/$(BIN)
install-debug: $(GOBIN)/$(BIN)-debug

$(GOBIN)/$(BIN): $(SRC)
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build $(GOOPTS) -o $@ $(REPO)/command/$(BIN)

$(GOBIN)/$(BIN)-debug: $(SRC)
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build $(GOOPTS) -tags debug -o $@ $(REPO)/command/$(BIN)


###
### clean
###

clean: clean-dist uninstall

uninstall:
	rm -rf $(GOBIN)/$(BIN)
	rm -rf $(GOBIN)/$(BIN)-debug

clean-dist:
	rm -rf dist

clean-docs:
	rm -rf docs/_site

###
### docs
###

docs:
	docker run --rm -v $(CWD)/docs:/site -p 4000:4000 andredumas/github-pages serve --watch


.PHONY: docs test clean