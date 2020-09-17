package tools

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"net"
	"net/url"
	"strings"
)

type ChangeType string

const (
	Add    ChangeType = "add"
	Update ChangeType = "update"
	Delete ChangeType = "delete"
)

const (
	question = "?"
)

func InitIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Printf("getIp fail,err is %v", err)
	}
	defer conn.Close()
	ip := strings.Split(conn.LocalAddr().String(), ":")[0]
	return ip
}

func StrMaxLimit(str string, max int) string {
	if len(str) > max {
		str = str[0:max]
		return str
	}
	return str
}

func SignString(stringToSign string, accessKeySecret string) string {
	key := []byte(accessKeySecret)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func Url2PathWithQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	pathWithQuery := u.Path

	if len(u.RawQuery) > 0 {
		pathWithQuery += question + u.RawQuery
	}
	return pathWithQuery
}
