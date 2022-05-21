# bloc-server
In a global view, bloc's code repo has below relationship:
![repo_relationship](/static/repo_relationship.png)

# about bloc-server
## responsibility
1. api server for bloc-frontend
2. scheduler for trigger flow、function's run, also include schedule retry、handle intercept、timeout...
3. store data

## relied on:
- [mongoDB](https://www.mongodb.com/): Database. Use to store functions、flow、run_record...'s infomation. Tested version: 5.0.5
- [rabbitMQ](https://www.rabbitmq.com/): MQ. Use to delivery trigger flow/function run msg. Tested version: 3.9.11
- [minio](https://github.com/minio/minio): Object storage. Used to store function run's output data. Tested version: RELEASE.2021-11-24T23-19-33Z
- [influxDB](https://github.com/influxdata/influxdb): Time series database. Used to store logs. Tested version: 2.1.1

## how to run bloc-server
> if you just want a local bloc environment（include both upper requirements、bloc-server、bloc-frontend）which can be used to receive your function's register and provide frontend ui. Just follow this [tutorial](https://fbloc.github.io/docs/deployGuide).

- after deployed above requirement, you can start up bloc-server by:
    ```shell
    $ go run cmd/server/main.go --app_name="$your_app_name" --rabbitMQ_connection_str="$rabbit_user:$rabbit_password@$rabbitMQ_server_address" --minio_connection_str="$minio_user:$minio_password@$mioio_server_address" --mongo_connection_str="$mongodb_user:$mongodb_password@$mongodb_address" --influxdb_connection_str="$influxDB_user:$influxDB_password@$influxDB_address/?token=$influxDB_token&organization=bloc
    ```
