package tools

const EmptyNotificationId = -1

type NamespaceData struct {
	NotificationId int               `json:"notificationId"`
	Configurations map[string]string `json:"configurations"`
}

type NotificationDto struct {
	NamespaceName  string `json:"namespaceName"`
	NotificationId int    `json:"notificationId"`
}