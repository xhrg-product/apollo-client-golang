package apollo

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xhrg-product/apollo-client-golang/no_ref"
	"github.com/xhrg-product/apollo-client-golang/tools"
	"io/ioutil"
	"log"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var nilMap = make(map[string]string)

const (
	httpHeaderAuthorization = "Authorization"
	httpHeaderTimestamp     = "Timestamp"
	authorizationFormat     = "Apollo %s:%s"
	delimiter               = "\n"
)

func NewClient(option *Options) *ApolloClient {
	//实例化client对象
	client := &ApolloClient{ConfigUrl: option.ConfigUrl, AppId: option.AppId, Cluster: option.Cluster, Secret: option.Secret}
	//赋值
	client.Secret = option.Secret
	client.Ip = tools.InitIp()
	//设置文件缓存路径
	if option.filePath == "" {
		client.CacheFilePath = no_ref.HomeDir() + "/data/apollo/cache/"
	} else {
		client.CacheFilePath = option.filePath
		if client.CacheFilePath[len(client.CacheFilePath)-1:] != "/" {
			client.CacheFilePath = client.CacheFilePath + "/"
		}
	}
	//启动热更新
	client.startHotUpdate()
	client.startHeartBeat()
	return client
}

type ApolloClient struct {
	//公开参数
	ConfigUrl     string
	Cluster       string
	AppId         string
	Ip            string
	CacheFilePath string
	Secret        string
	//私有参数
	cache          sync.Map
	changeListener func(changeType tools.ChangeType, namespace string, key string, value string)
	dataHash       sync.Map
	cycleTime      int
	stop           bool
	noKeyMap       sync.Map
}

func (client *ApolloClient) SetChangeListener(f func(changeType tools.ChangeType, namespace string, key string, value string)) {
	client.changeListener = f
}

func (client *ApolloClient) getFromNetV2(namespace string) *tools.NamespaceData {
	releaseKey := ""
	url := fmt.Sprintf("%s/configs/%s/%s/%s?releaseKey=%s&ip=%s", client.ConfigUrl, client.AppId, client.Cluster, namespace, releaseKey, client.Ip)
	code, body := tools.HttpRequest(url, 3, client.HTTPHeaders(url, client.AppId, client.Secret))
	if code != 200 {
		return nil
	}
	var x tools.UnCacheData
	json.Unmarshal([]byte(body), &x)
	data := &tools.NamespaceData{NotificationId: tools.EmptyNotificationId}
	data.Configurations = x.Configurations
	data.ReleaseKey = x.ReleaseKey
	return data
}

func (client *ApolloClient) GetValues(namespace string) map[string]string {
	if namespace == "" {
		namespace = "application"
	}
	//读取本地缓存
	if namespaceCache, ok := client.cache.Load(namespace); ok {
		kvData := namespaceCache.(*tools.NamespaceData).Configurations
		return kvData
	}
	//读取网络缓存
	namespaceNet := client.getFromNetV2(namespace)
	if namespaceNet != nil {
		kvData := namespaceNet.Configurations
		if kvData != nil {
			client.updateCache(namespace, namespaceNet)
			client.updateFile(namespace, namespaceNet)
			return kvData
		}
	}
	//读取文件缓存
	namespaceFile := client.getFileCache(namespace)
	if namespaceFile != nil {
		kvData := namespaceFile.Configurations
		if kvData != nil {
			client.updateCache(namespace, namespaceFile)
			return kvData
		}
	}
	client.setNilCache(namespace, "")
	return nilMap
}

//正常情况下，配置中心的key不可能为空的，
//这个方法对key为""做了判断，是因为根据namespace获取全部kv的时候，如果namespace不存在则把对应的缓存设置为空map。
//这个时候，不能map中包key为""，value为""的键值对
func (client *ApolloClient) setNilCache(namespace string, key string) {
	namespaceCache, ok := client.cache.Load(namespace)
	if !ok || namespaceCache == nil {
		data := &tools.NamespaceData{}
		maps := make(map[string]string)
		if key != "" {
			maps[key] = ""
		}
		data.Configurations = maps
		client.cache.Store(namespace, data)
		return
	}
	mData := namespaceCache.(*tools.NamespaceData)
	m := mData.Configurations
	if m == nil {
		maps := make(map[string]string)
		if key != "" {
			maps[key] = ""
		}
		mData.Configurations = maps
		return
	}
	if _, ok := m[key]; !ok {
		if key != "" {
			m[key] = ""
		}
		return
	}
}

func (client *ApolloClient) GetValue(key string, namespace string, defaultValue string) string {
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
	//读取不存在的key值，如果本地内存有不存在的key放入，则返回默认值。
	noneKey := client.noneKey(namespace, key)
	_, ok := client.noKeyMap.Load(noneKey)
	if ok {
		return defaultValue
	}
	//读取网络缓存
	namespaceNet := client.getFromNetV2(namespace)
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
	client.setNoKeyCache(namespace, key)
	return defaultValue
}

func (client *ApolloClient) noneKey(namespace string, key string) string {
	noneKey := fmt.Sprintf("%d%s%s", len(namespace), namespace, key)
	return noneKey
}

func (client *ApolloClient) setNoKeyCache(namespace string, key string) {
	nokey := client.noneKey(namespace, key)
	client.noKeyMap.Store(nokey, key)
}

func (client *ApolloClient) GetStringValue(key string, namespace string, defaultValue string) string {
	value := client.GetValue(key, namespace, defaultValue)
	return value
}

func (client *ApolloClient) GetBoolValue(key string, namespace string, defaultValue bool) bool {
	value := client.GetValue(key, namespace, "")
	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return b
}

func (client *ApolloClient) GetIntValue(key string, namespace string, defaultValue int) int {
	value := client.GetValue(key, namespace, "")
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}
func (client *ApolloClient) GetFloatValue(key string, namespace string, defaultValue float64) float64 {
	value := client.GetValue(key, namespace, "")
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

//func (client *ApolloClient) getFromNet(namespace string) *tools.NamespaceData {
//	url := fmt.Sprintf("%s/configfiles/json/%s/%s/%s?ip=%s", client.ConfigUrl, client.AppId, client.Cluster, namespace, client.Ip)
//	code, body := tools.HttpRequest(url, 3, client.HTTPHeaders(url, client.AppId, client.Secret))
//	if code != 200 {
//		return nil
//	}
//	namespaceData := &tools.NamespaceData{NotificationId: tools.EmptyNotificationId}
//	mapKv := make(map[string]string)
//	error := json.Unmarshal([]byte(body), &mapKv)
//	if error != nil {
//		return nil
//	}
//	namespaceData.Configurations = mapKv
//	return namespaceData
//}

func (client *ApolloClient) Stop() {
	client.stop = true
}

func (client *ApolloClient) startHotUpdate() {
	client.longPoll(false)
	go func() {
		for true {
			time.Sleep(time.Second * 3)
			if client.stop {
				return
			}
			client.longPoll(true)
		}
	}()
}

func (client *ApolloClient) startHeartBeat() {
	go func() {
		for true {
			time.Sleep(time.Minute * 10)
			if client.stop {
				return
			}
			client.cache.Range(func(key, value interface{}) bool {
				namespaceName := key
				namespaceData := value.(*tools.NamespaceData)
				url := fmt.Sprintf("%s/configs/%s/%s/%s?releaseKey=%s&ip=%s", client.ConfigUrl, client.AppId, client.Cluster, namespaceName, namespaceData.ReleaseKey, client.Ip)
				code, body := tools.HttpRequest(url, 3, client.HTTPHeaders(url, client.AppId, client.Secret))
				//如果返回code是200，则表示配置有变化了
				if code == 200 {
					var x tools.UnCacheData
					json.Unmarshal([]byte(body), &x)
					//触发回调接口：
					client.callListener(x.NamespaceName, namespaceData.Configurations /*旧的kv组合*/, x.Configurations /*新的kv组合*/)
					//给本地缓存namespaceData设置新的ReleaseKey和Configurations
					namespaceData.ReleaseKey = x.ReleaseKey
					namespaceData.Configurations = x.Configurations
					//这里不需要更新本地缓存，只需要更新文件缓存，因为namespaceData就是本地缓存。
					client.updateFile(x.NamespaceName, namespaceData)
				}
				return true
			})
		}
	}()
}

func (client *ApolloClient) longPoll(needChangeListener bool) {
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
	url := fmt.Sprintf("%s/notifications/v2?%s", client.ConfigUrl, paramStr)
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
		if len(array1) == 0 {
			return
		}
		for _, e := range array1 {
			namespaceName := e.NamespaceName
			data := client.getFromNetV2(namespaceName)
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
