package tools

const EmptyNotificationId = -1

type NamespaceData struct {
	NotificationId int               `json:"notificationId"`
	Configurations map[string]string `json:"configurations"`
	ReleaseKey     string            `json:"releaseKey"`
}

type NotificationDto struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int    `json:"notificationId"`
}

type CacheNode struct {
	//value指的是value
	Value string
	//entry指的是序列化的struct
	Entry interface{}
}

type UnCacheData struct {
	AppId          string            `json:"appId"`
	Cluster        string            `json:"cluster"`
	NamespaceName  string            `json:"namespaceName"`
	Configurations map[string]string `json:"configurations"`
	ReleaseKey     string            `json:"releaseKey"`
}
