# crontab 
1. 定时刷新公众号和企微接口调用凭据（access_token）
2. 提供接口用于访问公众号和企微的access_token

## 公众号access_token文档
1. 微信公众号测试号[申请地址](https://mp.weixin.qq.com/debug/cgi-bin/sandbox?t=sandbox/login) 
2. access_token获取[文档地址](https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Get_access_token.html)
## 企业微信access_token文档
1. access_token获取[文档地址](https://developer.work.weixin.qq.com/document/path/91039)

## crontab 开源项目
1. 开源项目地址：[robfig/cron](https://github.com/robfig/cron)
2. crontab 基本语法
```*     *     *   *    *        command to be executed
-     -     -   -    -
|     |     |   |    |
|     |     |   |    +----- day of the week (0 - 6) (Sunday = 0)
|     |     |   +------- month (1 - 12)
|     |     +--------- day of the month (1 - 31)
|     +----------- hour (0 - 23)
+------------- min (0 - 59)

```
各字段含义如下：
* min：表示分钟数，取值范围为 0-59
* hour：表示小时数，取值范围为 0-23
* day of the month：表示月份中的日期，取值范围为 1-31
* month：表示月份，取值范围为 1-12
* day of the week：表示星期几，取值范围为 0-6（其中 0 表示星期日）
* command to be executed：需要执行的命令或脚本路径
特殊字符含义如下：
* 星号（*）：匹配任意值
* 逗号（,）：可用于分隔多个取值
* 中划线（–）：可用于表示连续区间内的所有数值
* 斜线（/）：可用于表示每隔多长时间执行一次，例如 */5 表示每隔 5 分钟执行一次
例如：
``` 
# 该语句表示每隔两小时（整点开始）执行 /home/user/script.py 脚本
0 */2 * * * /usr/bin/python3 /home/user/script.py
```

# gRPC 代码生成
``` 
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative .\proto\token.proto
```

# docker build 
```
docker build -t crontab:0.1.0 .
```


# docker service 部署服务
## 创建配置文件资源
``` 
docker config create --label env=prod prod-crontab-conf config.yaml
```
## 创建并启动服务
``` 
docker service create --name crontab -p 50058:50058 \
--config src=prod-crontab-conf,target=/app/config.yaml \
--network 0voice-basic \
--replicas 1 --limit-cpu 0.3  \
--update-parallelism=1 \
--health-cmd "grpc_health_probe -addr=:50058" \
--health-interval 5s --health-retries 3 \
--with-registry-auth \
docker.0voice.com/crontab:0.1.0
```