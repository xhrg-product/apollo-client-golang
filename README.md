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
* 热更新


## 更多使用

* 见demo目录

## 注意点
 
