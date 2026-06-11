DOCKERVER :=`cat VERSION`
.DEFAULT_GOAL := build
VERSION :=`cat VERSION`

# Linux release binary (static, amd64)
gleaner-release:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gleaner ./cmd/gleaner/

build:
	go build -o gleaner ./cmd/gleaner/

test:
	go test ./pkg/config/ ./pkg/graph/

test-all:
	go test ./...

vet:
	go vet ./...

check: build vet test smoke
	@echo "All checks passed."

smoke: build
	./tools/smoke_test.sh ./gleaner

validate-config:
	@if [ -z "$(CFG_DIR)" ]; then echo "Usage: make validate-config CFG_DIR=configs/local"; exit 1; fi
	./tools/validate_config.sh $(CFG_DIR)

docker:
	podman build  --tag="fils/gleaner:$(VERSION)"  --file=./build/Dockerfile .

dockerpush:
	podman push localhost/fils/gleaner:$(VERSION) fils/gleaner:$(VERSION)
	podman push localhost/fils/gleaner:$(VERSION) fils/gleaner:latest

publish:
	docker tag fils/gleaner:$(VERSION) fils/gleaner:latest
	docker push fils/gleaner:$(VERSION) ; \
	docker push fils/gleaner:latest

releases: gleaner-release docker dockerpush publish

.PHONY: gleaner-release build test test-all vet check smoke validate-config docker dockerpush publish releases
