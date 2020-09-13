package apollo_client

import (
	"io/ioutil"
	"net/http"
	"time"
)

const errorCode = -1

func httpRequest(url string, timeout int) (int, string) {
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, errGet := client.Get(url)
	if errGet != nil {
		return errorCode, ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errorCode, ""
	}
	return resp.StatusCode, string(body)
}
