package tools

import (
	"github.com/xhrg-product/apollo-client-golang/a_log"
	"io/ioutil"
	"net/http"
	"time"
)

const ErrorCode = -1

func HttpRequest(url string, timeout int, header map[string][]string) (int, string) {

	timeStart := time.Now().Unix()
	a_log.Log().Infof("httpRequest start, url is %v", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		a_log.Log().Errorf("http.NewRequest error, url is %v, error is %v", url, err.Error())
		return ErrorCode, ""
	}
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	if header != nil {
		req.Header = header
	}
	resp, errGet := client.Do(req)

	timeEnd := time.Now().Unix()
	timeCostSecond := timeEnd - timeStart

	if errGet != nil {
		a_log.Log().Errorf("http request error, timeCostSecond is %v, error is %v", timeCostSecond, errGet)
		return ErrorCode, ""
	}
	if resp.StatusCode == 401 {
		a_log.Log().Errorf("http request error 401, timeCostSecond is %v, please check secret, url is %v", timeCostSecond, url)
		return ErrorCode, ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a_log.Log().Errorf("ioutil.ReadAll error ,error is %v", err)
		return ErrorCode, ""
	}
	bodyString := string(body)
	a_log.Log().Infof("HttpRequest end, uri is %v, code is %v, body is %v", url, resp.StatusCode, bodyString)
	return resp.StatusCode, bodyString

}
