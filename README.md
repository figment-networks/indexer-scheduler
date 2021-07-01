# Indexer Scheduler

Indexer scheduler is a simple continuous delay service, that allows to request services continuously.
Scheduler supports different kind of scenarios(runners).

## Configuration

By default scheduler runs on 8075. This port is for status, metrics and maintenance only.
As extra scheduler needs two additional config directories to be set.

### Schedules

Schedules config is responsible for passing different configurations options into tasks

```json
[{
    "network": "name",
    "chain_id": "name",
    "interval": "1s",
    "kind": "lastdata",
    "task_id": "unique_name"
}]
```

Parameters are self explanatory.
What is worth to mention task_id has to be set as unique string that should not be changed after initial setting.
Kind refers to runner name. currently only `lastdata` is supported

### Destinations

Destinations config refers to destination that scraper should respect
```json
[{
    "network": "name",
    "chain_id": "name",
    "version": "0.0.1",
    "conn_type": "http",
    "address": "http://0.0.0.0:8885",
    "additional": {
        "runner_name": {
            // additional params goes here
        }
    }
}]
```

## Local Development

### Get the code

```sh
git clone git@github.com:figment-networks/indexer-scheduler.git
cd indexer-scheduler/
```

### Dependencies

1. nodejs & npm (if you don't have it installed, see https://github.com/asdf-vm/asdf-nodejs)
2. [Go](https://golang.org/doc/install)
3. [Docker](https://www.docker.com/products/docker-desktop)

### Start services

```sh
docker-compose up -d postgresdatabase
```

Run database migrations. This is run separately to allow time for postgres to boot and be ready for migrations.

```sh
docker-compose up schedulermigrate
```

### Build the front end

> Important! You will need `nodejs` and `npm` installed in order to build the UI.

```sh
make prepare-ui-install-modules
make prepare-ui
```

### Run the scheduler

> There are pre-defined config files in `./config/development` with default values. However, these may not be suitable for your needs. Refer to the [Configuration](#configuration) section above.

```sh
DATABASE_URL=postgres://scheduler:scheduler@localhost:5431/scheduler?sslmode=disable \
AUTH_USER=dev \
AUTH_PASSWORD=dev \
DESTINATIONS_CONFIG=./config/development/destinations \
ADDRESS=127.0.0.1:8075 \
go run ./cmd/scheduler
```

If everything goes well, you should see the following logs:

```log
{"level":"info","time":"2021-06-29T17:22:24.495-0400","msg":"indexer-scheduler  (git: ) - built at "}
{"level":"info","time":"2021-06-29T17:22:24.496-0400","msg":"[DB] Connecting to database..."}
{"level":"info","time":"2021-06-29T17:22:24.511-0400","msg":"[DB] Ping successfull..."}
{"level":"info","time":"2021-06-29T17:22:24.511-0400","msg":"[Scheduler] Adding scheduler..."}
{"level":"error","time":"2021-06-29T17:22:24.517-0400","msg":"[Scheduler] Error during initial load of scheduler","error":"no rows updated"}
{"level":"info","time":"2021-06-29T17:22:24.517-0400","msg":"[Scheduler] Loading schedule initial config"}
{"level":"info","time":"2021-06-29T17:22:24.527-0400","msg":"[Scheduler] Loading destinations initial config from path"}
{"level":"info","time":"2021-06-29T17:22:24.528-0400","msg":"[Scheduler] Running Load"}
{"level":"info","time":"2021-06-29T17:22:24.528-0400","msg":"[API] Connecting to websocket ","host":"ws://127.0.0.1:8085/ws"}
{"level":"info","time":"2021-06-29T17:22:24.528-0400","msg":"[HTTP] Listening on 127.0.0.1:8075"}
```

By default, scheduler will try to start a websocket connection with a [manager](https://github.com/figment-networks/indexer-manager) running on port `0.0.0.0:8085`. If there is no manager running, it will retry until it succeeds.

Whenever you spin up a new worker that has registered with a manager (e.g [kava-worker](https://github.com/figment-networks/kava-worker)), scheduler will internally track the new target and the target will now be eligible to receive tasks. If multiple targets exist for the same network and chain, then scheduler will distribute tasks between all applicable targets (round-robin).

```log
{"level":"info","time":"2021-06-29T17:18:14.985-0400","msg":"[Scheduler] Adding destination config","connection_type":"ws","network":"kava","chain_id":"kava-7"}
```

### Add a task

To have the scheduler start coordinating data retrieval with the manager, you need to add a task. This can be done via the UI. 

1. Open the UI at [http://127.0.0.1:8075/ui/](http://127.0.0.1:8075/ui/)
2. Click New Task
3. Enter the task details. For example, if you are running a [kava-worker](https://github.com/figment-networks/kava-worker) for the kava-7 chain, use the following values:
  TaskId: kava-7-lastdata
  Type: lastdata
  Network: kava
  ChainID: kava-7
  Interval: 10s
4. Add the task and navigate back to the task list (you may need to refresh the page)
5. Scheduled tasks are disabled by default. In the task list, click "Enable Task" to start the task.

### Debug with VSCode

The `.vscode` directory contains a launch config to debug scheduler. To start debugging, open the Debug panel (⇧⌘D) and click the green arrow.



## Runners
### Last Data
Last data scenario/runner is sending next requests to given destination in given intervals.
Every request includes it's last known state that other party is suppose to start from.
Service keeps information about last state inside database.

The exchange between service and scheduler is based on two structures.
The `LatestDataRequest` sent from scheduler to the service, and  `LatestDataResponse` sent the other way around.


LatestDataRequest

| Name      | Type      | JSON          | Description                                                                                           |
| --------- | --------- | ------------- | ----------------------------------------------------------------------------------------------------- |
| Network   | string    | network       | Name of the network                                                                                   |
| ChainID   | string    | chain_id      | Name of the chain                                                                                     |
| Version   | string    | version       | Version of the request                                                                                |
| TaskID    | string    | task_id       | User defined ID of a task. Used if more than one process would be running against the same service.   |
| LastHash  | string    | last_hash     | If applies - last hash that was confirmed                                                             |
| LastEpoch | string    | last_epoch    | If applies - last epoch that was confirmed                                                            |
| LastHeight| uint64    | last_height   | If applies - last height that was confirmed                                                           |
| LastTime  | time.Time | last_time     | If applies - last time that was confirmed                                                             |
| RetryCount| uint64    | retry_count   | Current time retry count                                                                              |
| Nonce     | []byte    | nonce         | Nonce, any information that should be passed back in next request that doesn't fit in above           |


LatestDataResponse

| Name      | Type      | JSON          | Description                                                                                           |
| --------- | --------- | ------------- | ----------------------------------------------------------------------------------------------------- |
| LastHash  | string    | last_hash     | If applies - last hash that was confirmed                                                             |
| LastEpoch | string    | last_epoch    | If applies - last epoch that was confirmed                                                            |
| LastHeight| uint64    | last_height   | If applies - last height that was confirmed                                                           |
| LastTime  | time.Time | last_time     | If applies - last time that was confirmed                                                             |
| RetryCount| uint64    | retry_count   | Current time retry count                                                                              |
| Nonce     | []byte    | nonce         | Nonce, any information that should be passed back in next request that doesn't fit in above           |
| Error     | []byte    | error         | Information about error during process.                                                               |
| Processing| bool      | processing    | True, if task is still processing. In that case backoff strategy will be used                         |


If used with http transport runner needs additional configuration:
```json
    {
        "endpoint": "/scrape_latest"
    }
```

Where `endpoint` is the endpoint compatible with last data format, starting with preceding `/`
