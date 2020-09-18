package apollo

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xhrg-product/apollo-client-golang/tools"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	httpHeaderAuthorization = "Authorization"
	httpHeaderTimestamp     = "Timestamp"
	authorizationFormat     = "Apollo %s:%s"
	delimiter               = "\n"
)

func NewClient(option *Options) *ApolloClient {
	//实例化client对象
	client := &ApolloClient{ConfigServerUrl: option.ApolloConfigUrl, AppId: option.AppID, Cluster: option.Cluster}
	//赋值
	client.Secret = option.Secret
	client.Ip = tools.InitIp()
	client.CacheFilePath = tools.HomeDir() + "/data/apollo/cache/"
	//启动热更新
	client.startHotUpdate()
	return client
}

type ApolloClient struct {
	//公开参数
	ConfigServerUrl string
	Cluster         string
	AppId           string
	Ip              string
	CacheFilePath   string
	Secret          string

	//私有参数
	cache          sync.Map
	changeListener func(changeType tools.ChangeType, namespace string, key string, value string)
	dataHash       sync.Map
	cycleTime      int
	stop           bool
}

func (client *ApolloClient) SetChangeListener(f func(changeType tools.ChangeType, namespace string, key string, value string)) {
	client.changeListener = f
}

func (client *ApolloClient) GetValue(key string, namespace string) string {
	if namespace == "" {
		namespace = "application"
	}
	//读取本地缓存
	if namespaceCache, ok := client.cache.Load(namespace); ok {
		kvData := namespaceCache.(*tools.NamespaceData).Configurations
		if val, ok := kvData[key]; ok {
			return val
		}
	}

	//读取网络缓存
	namespaceNet := client.getFromNet(namespace)
	if namespaceNet != nil {
		kvData := namespaceNet.Configurations
		if kvData != nil {
			if val, ok := kvData[key]; ok {
				client.updateCache(namespace, namespaceNet)
				client.updateFile(namespace, namespaceNet)
				return val
			}
		}
	}

	//读取文件缓存
	namespaceFile := client.getFileCache(namespace)
	if namespaceFile != nil {
		kvData := namespaceFile.Configurations
		if kvData != nil {
			if val, ok := kvData[key]; ok {
				client.updateCache(namespace, namespaceFile)
				return val
			}
		}
	}
	client.setNilCache(namespace, key)
	return ""
}

func (client *ApolloClient) setNilCache(namespace string, key string) {
	namespaceCache, ok := client.cache.Load(namespace)
	if !ok || namespaceCache == nil {
		data := &tools.NamespaceData{}
		maps := make(map[string]string)
		maps[key] = ""
		data.Configurations = maps
		client.cache.Store(namespace, data)
		return
	}
	mData := namespaceCache.(*tools.NamespaceData)
	m := mData.Configurations
	if m == nil {
		maps := make(map[string]string)
		maps[key] = ""
		mData.Configurations = maps
		return
	}
	if _, ok := m[key]; !ok {
		m[key] = ""
		return
	}
}

func (client *ApolloClient) GetStringValue(key string, namespace string, defaultValue string) string {
	value := client.GetValue(key, namespace)
	if value == "" {
		return defaultValue
	}
	return value
}

func (client *ApolloClient) GetBoolValue(key string, namespace string, defaultValue bool) bool {
	value := client.GetValue(key, namespace)
	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return b
}

func (client *ApolloClient) GetIntValue(key string, namespace string, defaultValue int) int {
	value := client.GetValue(key, namespace)
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}
func (client *ApolloClient) GetFloatValue(key string, namespace string, defaultValue float64) float64 {
	value := client.GetValue(key, namespace)
	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

func (client *ApolloClient) getFileCache(namespace string) *tools.NamespaceData {
	filePath := client.filePath(namespace)
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil
	}
	namespaceData := &tools.NamespaceData{}
	err = json.Unmarshal(bytes, namespaceData)
	if err != nil {
		return nil
	}
	return namespaceData
}

//更新本地缓存
func (client *ApolloClient) updateCache(namespace string, data *tools.NamespaceData) {
	if namespace == "" || data == nil {
		return
	}
	if data.NotificationId != tools.EmptyNotificationId {
		client.cache.Store(namespace, data)
		return
	}
	nid := data.NotificationId
	if d, ok := client.cache.Load(namespace); ok {
		nd := d.(*tools.NamespaceData)
		if nd.NotificationId > 0 {
			nid = nd.NotificationId
		}
	}
	data.NotificationId = nid
	client.cache.Store(namespace, data)
}

func (client *ApolloClient) updateFile(namespace string, data *tools.NamespaceData) {

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

func (client *ApolloClient) filePath(namespace string) string {
	return fmt.Sprintf("%s%s_configuration_%s.txt", client.CacheFilePath, client.AppId, namespace)
}

func (client *ApolloClient) getFromNet(namespace string) *tools.NamespaceData {

	url := fmt.Sprintf("%s/configfiles/json/%s/%s/%s?ip=%s", client.ConfigServerUrl, client.AppId, client.Cluster, namespace, client.Ip)
	code, body := tools.HttpRequest(url, 3, client.HTTPHeaders(url, client.AppId, client.Secret))
	if code != 200 {
		return nil
	}
	namespaceData := &tools.NamespaceData{NotificationId: tools.EmptyNotificationId}
	mapKv := make(map[string]string)
	error := json.Unmarshal([]byte(body), &mapKv)
	if error != nil {
		return nil
	}
	namespaceData.Configurations = mapKv
	return namespaceData
}

func (client *ApolloClient) startHotUpdate() {

	client.hotUpdate(false)

	go func() {
		for true {
			time.Sleep(time.Second * 3)
			if client.stop {
				return
			}
			client.hotUpdate(true)
		}
	}()
}

func (client *ApolloClient) hotUpdate(needChangeListener bool) {

	var array []*tools.NotificationDto
	client.cache.Range(func(key, value interface{}) bool {
		if value != nil {
			array = append(array, &tools.NotificationDto{NamespaceName: key.(string), NotificationId: value.(*tools.NamespaceData).NotificationId})
		}
		return true
	})

	if len(array) == 0 {
		return
	}

	bytes, err := json.Marshal(array)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Add("appId", client.AppId)
	params.Add("cluster", client.Cluster)
	params.Add("notifications", string(bytes))
	paramStr := params.Encode()
	url := fmt.Sprintf("%s/notifications/v2?%s", client.ConfigServerUrl, paramStr)
	code, body := tools.HttpRequest(url, 75, client.HTTPHeaders(url, client.AppId, client.Secret))
	if code == 304 {
		return
	}
	if code == tools.ErrorCode {
		return
	}
	if code == 401 {
		logrus.Printf("http request error 401,please check secret")
		log.Printf("http request error 401,please check secret")
	}
	if code == 200 {
		var array1 []*tools.NotificationDto
		err := json.Unmarshal([]byte(body), &array1)
		if err != nil {
			log.Println(err)
		}
		//[{"namespaceName":"application","notificationId":974,"messages":{"details":{"demo-service+default+application":974}}}]
		if len(array1) == 0 {
			return
		}
		for _, e := range array1 {
			namespaceName := e.NamespaceName
			data := client.getFromNet(namespaceName)
			if data == nil {
				continue
			}
			data.NotificationId = e.NotificationId

			//调用回调函数
			if needChangeListener {
				mapkvOld := make(map[string]string)
				if namespaceCache, ok := client.cache.Load(namespaceName); ok {
					mapkvOld = namespaceCache.(*tools.NamespaceData).Configurations
				}
				client.callListener(namespaceName, mapkvOld, data.Configurations)
			}

			client.updateCache(namespaceName, data)
			client.updateFile(namespaceName, data)
		}
	}

}

func (client *ApolloClient) callListener(namespace string, oldKv map[string]string, newKv map[string]string) {
	if client.changeListener == nil {
		return
	}
	if oldKv == nil {
		oldKv = make(map[string]string)
	}
	if newKv == nil {
		newKv = make(map[string]string)
	}
	for oldk, oldv := range oldKv {
		newv, ok := newKv[oldk]
		if !ok {
			client.changeListener(tools.Delete, namespace, oldk, oldv)
			continue
		}
		if newv != oldv {
			client.changeListener(tools.Update, namespace, oldk, newv)
			continue
		}
	}
	for newk, newv := range newKv {
		_, ok := oldKv[newk]
		if !ok {
			client.changeListener(tools.Add, namespace, newk, newv)
			continue
		}
	}
}

// HTTPHeaders HTTPHeaders
func (client *ApolloClient) HTTPHeaders(url string, appID string, secret string) map[string][]string {
	if secret == "" {
		return nil
	}
	ms := time.Now().UnixNano() / int64(time.Millisecond)
	timestamp := strconv.FormatInt(ms, 10)
	pathWithQuery := tools.Url2PathWithQuery(url)

	//正常的path一般是 /query 但是可能会出现 //query
	if pathWithQuery != "" && len(pathWithQuery) > 2 {
		a := pathWithQuery[0:2]
		if a == "//" {
			pathWithQuery = pathWithQuery[1:len(pathWithQuery)]
		}
	}
	stringToSign := timestamp + delimiter + pathWithQuery
	signature := tools.SignString(stringToSign, secret)
	headers := make(map[string][]string, 2)
	signatures := make([]string, 0, 1)
	signatures = append(signatures, fmt.Sprintf(authorizationFormat, appID, signature))
	headers[httpHeaderAuthorization] = signatures
	timestamps := make([]string, 0, 1)
	timestamps = append(timestamps, timestamp)
	headers[httpHeaderTimestamp] = timestamps
	return headers
}
