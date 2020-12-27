.phony: build
build:
	@go build -v -o /tmp/pcloud ./cmd

.phony: test
test: test-sdk test-tracker

test-sdk:
	@[ -n "$$GO_PCLOUD_USERNAME" ] || read -p "user? " GO_PCLOUD_USERNAME ; \
	 [ -n "$$GO_PCLOUD_PASSWORD" ] || { read -s -p "pass? " GO_PCLOUD_PASSWORD && echo; } ; \
	 [ -n "$$GO_PCLOUD_TFA_CODE" ] || { read -s -p "tfa code? " GO_PCLOUD_TFA_CODE && echo; } ; \
	 GO_PCLOUD_USERNAME="$$GO_PCLOUD_USERNAME" GO_PCLOUD_PASSWORD="$$GO_PCLOUD_PASSWORD" GO_PCLOUD_TFA_CODE="$$GO_PCLOUD_TFA_CODE" go test -v -count 1 -race -timeout 10s ./sdk

test-tracker:
	@go test -v -count 1 -race -timeout 10s ./tracker

.phony: lint
lint:
	@which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.33.0
	@golangci-lint run ./...


