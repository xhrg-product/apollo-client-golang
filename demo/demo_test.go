package demo

import (
	"github.com/sirupsen/logrus"
	"github.com/xhrg-product/apollo-client-golang/apollo"
	"os"
	"testing"
)

func TestGetString(t *testing.T) {
	configUrl := os.Getenv("APOLLO_CONFIG_URL")
	apolloClient := apollo.NewClient(&apollo.Options{ConfigUrl: configUrl, AppId: "demo-service", Cluster: "default"})
	val := apolloClient.GetStringValue("name", "application", "defaultValue")
	logrus.Info(val)
}

func TestGetInt(t *testing.T) {
	configUrl := os.Getenv("APOLLO_CONFIG_URL")
	apolloClient := apollo.NewClient(&apollo.Options{ConfigUrl: configUrl, AppId: "demo-service", Cluster: "default"})
	val := apolloClient.GetIntValue("name", "application", 10)
	logrus.Info(val)
}
