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
