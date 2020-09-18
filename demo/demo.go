package demo

import (
	"github.com/sirupsen/logrus"
	"github.com/xhrg-product/apollo-client-golang/apollo"
	"os"
	"testing"
)

func TestC(t *testing.T) {
	configUrl := os.Getenv("APOLLO_CONFIG_URL")
	apolloClient := apollo.NewClient(&apollo.Options{ApolloConfigUrl: configUrl, AppID: "demo-service", Cluster: "default"})
	val := apolloClient.GetStringValue("name", "application", "defaultValue")
	logrus.Info(val)
}
