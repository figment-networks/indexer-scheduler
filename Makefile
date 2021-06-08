
LDFLAGS      := -w -s
MODULE       := github.com/figment-networks/indexer-scheduler
VERSION_FILE ?= ./VERSION


# Git Status
GIT_SHA ?= $(shell git rev-parse --short HEAD)

ifneq (,$(wildcard $(VERSION_FILE)))
VERSION ?= $(shell head -n 1 $(VERSION_FILE))
else
VERSION ?= n/a
endif

all: build

.PHONY: build
build:
	$(info building scheduler binary as ./scheduler)
	go build -o scheduler ./cmd/scheduler

.PHONY: build-migration
build-migration:
	$(info building migration binary as ./migration)
	go build -o migration ./cmd/scheduler-migration

.PHONY: pack-release
pack-release:
	$(info preparing release)
	@mkdir -p ./release
	@make build-migration
	@mv ./migration ./release/migration
	@make build
	@mv ./scheduler ./release/scheduler
	@cp -R ./cmd/scheduler-migration/migrations ./release/
	@zip -r indexer-scheduler ./release
	@rm -rf ./release

.PHONY: generate
generate:
	go generate ./...

.PHONY: prepare-ui-install-modules
prepare-ui-install-modules:
	set NODE_ENV=development
	cd ./assets; npm install


.PHONY: prepare-ui
prepare-ui:
	set NODE_ENV=development
	cd ./assets; npm run build;
	mkdir -p ./ui
	rm -rf ./ui/assets
	mkdir ./ui/assets/
	cp -r ./assets/build/* ./ui/assets/
