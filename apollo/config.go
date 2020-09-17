package apollo

import (
	"github.com/sirupsen/logrus"
)

func InitLog(level logrus.Level) {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(level)
}

//AppConfig 配置文件
type Options struct {
	ApolloConfigUrl string
	AppID           string
	Cluster         string
	Secret          string
}
