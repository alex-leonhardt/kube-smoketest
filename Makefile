.PHONY: help build run clean

help:
	@echo "Options:"
	@echo "- build: build the binary, but don't run it."
	@echo "- run:   build and run the binary."
	@echo "- clean: clean up after, i.e. deletes the kube-smoketest namespace"

build:
	@echo "Building kube-smoketest binary.."
	@go build -o build/kube-smoketest

run: build
	@echo "Running kube-smoketest.."
	@build/kube-smoketest

clean:
	@kubectl delete namespace kube-smoketest
