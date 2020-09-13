package apollo_client

import (
	"log"
	"os"
	"testing"
)

func TestIp(t *testing.T) {

	namespaceData := &namespaceData{}
	namespaceData.Configurations["a"] = "a"
}

func TestHttp(t *testing.T) {
	code, body := httpRequest("http://www.baidu.comc", 4)
	println(code)
	println(body)
}

func TestNewApolloClient(t *testing.T) {
	url := os.Getenv("APOLLO_CONFIG_URL")
	log.Println(url)
	client := NewApolloClient(url, "demo-service", "DEV")
	client.GetValue("name", "", "")
}
