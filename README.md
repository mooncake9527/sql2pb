# sql2pb

根据mysql的表，生成proto文件

有哪些表需要忽略的，在 config/ignore.example.yaml 里面配置

```shell
go build .

./sql2pb proto -s tusihao:123456@192.168.0.188:3306 -d yl_company -c config/ignore.example.yaml -o pb/company/proto -t company,company_user

or

./sql2pb proto --server=tusihao:123456@192.168.0.188:3306 \
--db=yl_company \
--config=config/ignore.example.yaml \
--out=pb/company/proto \
--tables=company,company_user


```
