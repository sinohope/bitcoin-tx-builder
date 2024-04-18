tag := $(shell git describe --tags --exact-match 2>/dev/null || echo "")
commit := $(shell git rev-parse HEAD)
build_time := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MODULE = "gitlab.newhuoapps.com/dcenter/mpc-service"

LD_FLAGS := -ldflags "-w -s \
	-X gitlab.newhuoapps.com/dcenter/mpc-service/common.Tag=${tag} \
	-X gitlab.newhuoapps.com/dcenter/mpc-service/common.Commit=${commit} \
	-X gitlab.newhuoapps.com/dcenter/mpc-service/common.BuildTime=${build_time}"
CROSS_COMPILE = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
CILINT := $(shell command -v golangci-lint 2> /dev/null)
GOIMPORTS := $(shell command -v goimports 2> /dev/null)

style:
ifndef GOIMPORTS
	$(error "goimports is not available please install goimports")
endif
	! find . -path ./vendor -prune -o -name '*.go' -print | xargs goimports -d -local ${MODULE} | grep '^'

format:
ifndef GOIMPORTS
	$(error "goimports is not available please install goimports")
endif
	find . -path ./vendor -prune -o -name '*.go' -print | xargs goimports -l -local ${MODULE} | xargs goimports -l -local ${MODULE} -w

clean:
	rm -fr ./bitcoin-tx-builder



build: clean
	go build

.PHONY: build clean

