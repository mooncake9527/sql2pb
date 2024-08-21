# sql2pb

### 简介
一个从mysql表结构生成protobuf文件的小工具,使用配置文件来指定表结构和生成的protobuf文件的输出目录等信息


### 使用方法

#### 方法一:
安装
```shell
go install github.com/mooncake9527/sql2pb
```

使用shell
```shell
sql2pb proto -c config/config.yaml
```

#### 方法二:

使用go
```shell

git clone https://github.com/mooncake9527/sql2pb.git

cd sql2pb
go build -o sql2pb
./sql2pb proto -c config/config.yaml

```


### 配置文件说明 

config/config.yaml
```text
out: ./out  # 生成文件的输出目录
tpl: ./template/proto.tpl # 模板文件
db:
  host: localhost # 数据库地址
  port: 3306  # 数据库端口
  user: root  # 数据库用户名
  password: xxxxx   # 数据库密码
  schema: user  # 数据库名称
  tables: country,user  # 数据库表,多个用逗号分隔
```