.PHONY: help generate manifests run tidy

help:
	@echo "Available targets: generate, manifests, run, tidy"

# generate and manifests are stubbed because controller-gen is unavailable offline.
generate:
	@echo "Skipping code generation (controller-gen not available in offline environment)."

manifests:
	@echo "Skipping manifest generation (controller-gen not available in offline environment)."

tidy:
	go fmt ./...

run:
	go run ./main.go
