BUILD_TAGS?=paradigm

# vendor uses Glide to install all the Go dependencies in vendor/
vendor:
	glide install

# install compiles and places the binary in GOPATH/bin
install:
	go install --ldflags '-extldflags "-static"' \
		--ldflags "-X github.com/paradigm-network/paradigm/version.GitCommit=`git rev-parse HEAD`" \
		./cmd/paradigm

# build compiles and places the binary in /build
build:
	CGO_ENABLED=0 go build \
		--ldflags "-X github.com/paradigm-network/paradigm/version.GitCommit=`git rev-parse HEAD`" \
		-o build/paradigm ./cmd/paradigm/

test:
	glide novendor | xargs go test

.PHONY: vendor install build dist test
