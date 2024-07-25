# sql2pb

根据mysql的表，生成proto文件

有哪些表需要忽略的，在 config/ignore.example.yaml 里面配置

```shell
go build .

./sql2pb proto -s root:123456@localhost:3306 -d user -c config/ignore.example.yaml -o pb/user/proto -t user,user_info

or

./sql2pb proto --server=root:123456@localhost:3306 \
--db=user \
--config=config/ignore.example.yaml \
--out=pb/user/proto \
--tables=user,user_info


```
