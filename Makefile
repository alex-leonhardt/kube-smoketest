.PHONY: help build run debug clean

help:
	@echo "Options:"
	@echo "- build: build the binary, but don't run it."
	@echo "- run:   build and run the binary."
	@echo "- debug: build and run the binary with debug flag set."
	@echo "- clean: clean up after, i.e. deletes the kube-smoketest namespace"

build:
	@echo "Building kube-smoketest binary.."
	@go build -o build/kube-smoketest

run: build
	@echo "Running kube-smoketest.."
	@build/kube-smoketest -alsologtostderr -v=10

debug: build
	@echo "Running kube-smoketest w/ debug.."
	@build/kube-smoketest -alsologtostderr -v=10 -debug

clean:
	@kubectl delete namespace kube-smoketest
