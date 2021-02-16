.PHONY: all test validate fmt

export GOPROXY=https://proxy.golang.org

# BUILDTAGS = btrfs_noversion libdm_no_deferred_remove
# BUILDFLAGS := -tags "$(BUILDTAGS)"
BUILDFLAGS := ""

all: test validate

build:
	GO111MODULE="on" go build $(BUILDFLAGS) ./...

test:
	GO111MODULE="on" go test $(BUILDFLAGS) -cover ./...

fmt:
	@gofmt -l -s -w .

validate:
	@GO111MODULE="on" go vet ./...
	@test -z "$$(gofmt -s -l . | tee /dev/stderr)"
