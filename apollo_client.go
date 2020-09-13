package apollo_client

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

func NewApolloClient(configServerUrl string, appId string, cluster string) *apolloClint {
	client := &apolloClint{ConfigServerUrl: configServerUrl, AppId: appId, Cluster: cluster}
	client.Ip = initIp()
	client.CacheFilePath = HomeDir() + "/data/apollo/cache/"
	return client
}

type apolloClint struct {
	//公开参数
	ConfigServerUrl string
	Cluster         string
	AppId           string
	Ip              string
	CacheFilePath   string
	ChangeListener  func(changeType ChangeType, nammespace string, key string, value string)
	Secret          string

	//私有参数
	cache     sync.Map
	dataHash  sync.Map
	cycleTime int
}

func (client apolloClint) GetValue(key string, namespace string, defaultVal string) string {
	if namespace == "" {
		namespace = "application"
	}
	//读取本地缓存
	if namespaceCache, ok := client.cache.Load(key); ok {
		kvData := namespaceCache.(namespaceData).Configurations
		if val, ok := kvData[key]; ok {
			return val.(string)
		}
	}
	//读取文件缓存
	namespaceFile := client.getFileCache(namespace)
	if namespaceFile != nil {
		kvData := namespaceFile.Configurations
		if kvData != nil {
			if val, ok := kvData[key]; ok {
				client.updateCache(namespace, namespaceFile)
				return val.(string)
			}
		}
	}

	namespaceNet := client.getFromNet(namespace)
	if namespaceNet != nil {
		kvData := namespaceNet.Configurations
		if kvData != nil {
			if val, ok := kvData[key]; ok {
				client.updateCache(namespace, namespaceNet)
				client.updateFile(namespace, namespaceNet)
				return val.(string)
			}
		}
	}
	return defaultVal
}

func (client apolloClint) getFileCache(namespace string) *namespaceData {
	filePath := client.filePath(namespace)
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil
	}
	namespaceData := &namespaceData{}
	err = json.Unmarshal(bytes, namespaceData)
	if err != nil {
		return nil
	}
	return namespaceData
}

//更新本地缓存
func (client apolloClint) updateCache(namespace string, data *namespaceData) {
	if namespace == "" || data == nil {
		return
	}
	nid := data.NotificationId
	if d, ok := client.cache.Load(namespace); ok {
		nd := d.(namespaceData)
		if nd.NotificationId != "" {
			nid = nd.NotificationId
		}
	}
	data.NotificationId = nid
	client.cache.Store(namespace, data)
}

func (client apolloClint) updateFile(namespace string, data *namespaceData) {

	bytes, error := json.Marshal(data)
	if error != nil {
		return
	}
	jsonMd5 := fmt.Sprintf("%x", md5.Sum(bytes))
	if hash, ok := client.dataHash.Load(namespace); ok {
		if hash == jsonMd5 {
			return
		}
	}
	client.dataHash.Store(namespace, jsonMd5)
	filePath := client.filePath(namespace)
	ioutil.WriteFile(filePath, bytes, 0666)
}

func (client apolloClint) filePath(namespace string) string {
	return fmt.Sprintf("%s%s_configuration_%s.txt", client.CacheFilePath, client.AppId, namespace)
}

func (client apolloClint) getFromNet(namespace string) *namespaceData {

	url := fmt.Sprintf("%s/configfiles/json/%s/%s/%s?ip=%s", client.ConfigServerUrl, client.AppId, client.Cluster, namespace, client.Ip)
	_, body := httpRequest(url, 3)

	namespaceData := &namespaceData{NotificationId: "0"}
	mapKv := make(map[string]interface{})

	error := json.Unmarshal([]byte(body), &mapKv)
	if error != nil {
		return nil
	}
	namespaceData.Configurations = mapKv
	return namespaceData
}
