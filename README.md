# fansMedal

## 配置文件

文件名：`config.yaml`

```yaml
list:
  - 1 # 过滤名单
type: false # 为true 时，list为白名单，否则为黑名单
accessKey: "77379e48c66bf45e7xxxxx"
likeCD: 3
chatCD: 5
isWear: true
```

## 编译

```shell
git clone https://github.com/Hami-Lemon/fansMedal.git
cd fansMedal/cmd/medal
go build .
```