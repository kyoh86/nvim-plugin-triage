VERSION ?= `git vertag get`
COMMIT  ?= `git rev-parse HEAD`
DATE    ?= `date --iso-8601`

lint:
	go tool golangci-lint run
.PHONY: lint

test:
	go test -v --race ./...
.PHONY: test

install: test
	go install -a -ldflags "-X=main.version=$(VERSION) -X=main.commit=$(COMMIT) -X=main.date=$(DATE)" ./cmd/nvim-plugin-triage/...
.PHONY: install

default: lint test
.DEFAULT_GOAL := default
