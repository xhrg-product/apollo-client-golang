package tools

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

const ErrorCode = -1

func HttpRequest(url string, timeout int, header map[string][]string) (int, string) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ErrorCode, ""
	}

	client := http.Client{Timeout: time.Duration(timeout) * time.Second}

	if header != nil {
		req.Header = header
	}

	resp, errGet := client.Do(req)

	if errGet != nil {
		logrus.Printf("http request error, error is %s", errGet)
		return ErrorCode, ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ErrorCode, ""
	}
	return resp.StatusCode, string(body)
}
