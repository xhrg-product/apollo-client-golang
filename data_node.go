package apollo_client

type namespaceData struct {
	NotificationId string `json:"notificationId"`
	//这里的map中value必须是interface{}，因为golang的字符串不支持nil赋值，
	//请求传递一个错误的key，没有获取到key对应的value，第二次请求应直接按照nil处理，
	//所以本地缓存需要填充一个nil，而不是空字符串，但是apollo的配置是支持空字符串的
	//如果填充空字符串和apollo配置的空字符串相互违背，所以这里用interface{}是为了填充缓存nil
	Configurations map[string]interface{} `json:"configurations"`
}
