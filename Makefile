DOCKERVER :=`cat VERSION`
.DEFAULT_GOAL := nabu
VERSION :=`cat VERSION`

nabu:
	cd cmd/nabu; \
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o nabu

build:
	go build -o nabu ./cmd/nabu/

test:
	go test ./pkg/config/ ./pkg/graph/

test-all:
	go test ./...

vet:
	go vet ./...

check: build vet test smoke
	@echo "All checks passed."

smoke: build
	./tools/smoke_test.sh ./nabu

validate-config:
	@if [ -z "$(CFG_DIR)" ]; then echo "Usage: make validate-config CFG_DIR=configs/local"; exit 1; fi
	./tools/validate_config.sh $(CFG_DIR)

docker:
	podman build  --tag="fils/nabu:$(VERSION)"  --file=./build/Dockerfile .

dockerpush:
	podman push localhost/fils/nabu:$(VERSION) fils/nabu:$(VERSION)
	podman push localhost/fils/nabu:$(VERSION) fils/nabu:latest

publish:
	docker tag fils/nabu:$(VERSION) fils/nabu:latest
	docker push fils/nabu:$(VERSION) ; \
	docker push fils/nabu:latest

releases: nabu docker dockerpush publish

.PHONY: nabu build test test-all vet check smoke validate-config docker dockerpush publish releases
