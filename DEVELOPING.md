# Setup

## Get the code

```sh
git clone git@github.com:figment-networks/indexer-scheduler.git
cd indexer-scheduler/
```

## Start services

```sh
docker-compose up -d postgresdatabase
```

Run database migrations. This is run separately to allow time for postgres to boot and be ready for migrations.

```sh
docker-compose up schedulermigrate
```

## Build the front end

> Important! You will need `nodejs` and `npm` installed in order to build the UI. If you don't have this installed, check out https://github.com/asdf-vm/asdf-nodejs.

```
make prepare-ui-install-modules
make prepare-ui
```

## Run the scheduler

```sh
source config/development/env
go run cmd/scheduler/main.go cmd/scheduler/dynamic.go cmd/scheduler/profiling.go
```

## View the UI

```
open http://127.0.0.1:8075/ui/
```