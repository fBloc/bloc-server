# bloc-server
bloc的各个代码仓库的关系如下：
![repo_relationship](/static/repo_relationship.png)

## 职责
1. bloc-frontend 的 http api server
2. 调度器，负责触发flow以及function的运行
3. 存储数据、维护状态

## 依赖项
- [mongoDB](https://www.mongodb.com/): 数据库. 用于存储function、flow、运行记录等的信息. 版本 5.0.5 已测试
- [rabbitMQ](https://www.rabbitmq.com/): 消息队列. 用于发布flow/function的运行消息. 版本 3.9.11 已测试
- [minio](https://github.com/minio/minio): 对象存储. 用于存储函数运行的输出数据（因为如果直接使用mongo存储输出、在遇到某个输出值是很大的数据时，会可能无法支撑）。版本 RELEASE.2021-11-24T23-19-33Z 已测试
- [influxDB](https://github.com/influxdata/influxdb): 时序数据库. 用于存储日志. 版本 2.1.1 已测试

## how to run bloc-server
> 如果你只是想部署个本地测试环境(包含上面的依赖项、bloc-server、bloc-frontend) 用于接收你开发的bloc function并提供bloc web访问端，请按照此[教程]((https://fbloc.github.io/docs/deployGuide))

- 在部署好了上面的依赖项目后，你可以通过下面的命令启动bloc-server:
    ```shell
    $ go run cmd/server/main.go --app_name="$your_app_name" --rabbitMQ_connection_str="$rabbit_user:$rabbit_password@$rabbitMQ_server_address" --minio_connection_str="$minio_user:$minio_password@$mioio_server_address" --mongo_connection_str="$mongodb_user:$mongodb_password@$mongodb_address" --influxdb_connection_str="$influxDB_user:$influxDB_password@$influxDB_address/?token=$influxDB_token&organization=bloc
    ```
