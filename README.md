## apollo-client-golang


## 使用

```golang
在go.mod配置
require (
	github.com/xhrg-product/apollo-client-golang v1.0.2 //请确认使用最新tag
)

main.go中代码如下
configUrl := os.Getenv("APOLLO_CONFIG_URL")
apolloClient := apollo.NewClient(&apollo.Options{ConfigUrl: configUrl, AppId: "demo-service", Cluster: "default"})
val := apolloClient.GetStringValue("name", "application", "defaultValue")
logrus.Info(val)
```

## 功能点

* 拉取apollo配置中心的配置。namespace,key
* 本地文件缓存
* 错误的key和namespace特殊处理
* secret认证
* 使用获取数据库的接口热更新【如果使用缓存接口,当config节点A通知更新，而节点2并没有更新到的时候，会获取不到更新】
* 心跳机制，如果没有心跳，apollo会看不到实例。
* 日志。在业务控制台只打印错误日志。info日志会放在{userhome}/data/logs下，因为info日志1分钟1次，所以无需关闭。

## 更多使用

* 见demo目录

## 注意点
 
