PARADIGM_VERSION=0.1

all: paradigm-image

paradigm-image:
	go build \
		--ldflags '-extldflags "-static"' \
		-o paradigm/paradigm ../cmd/paradigm/
	docker build -t paradigm-network/paradigm:$(PARADIGM_VERSION) \
		-t paradigm-network/paradigm:latest \
		paradigm/

	
.PHONY: all paradigm-image