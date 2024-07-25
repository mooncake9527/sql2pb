# sql2pb

A tool for generating proto files from mysql

install
```shell
go install github.com/mooncake9527/sql2pb
```

generate proto files
```shell
sql2pb proto -s root:123456@localhost:3306 -d user -o pb/user/proto -t user,user_info
```
