package a_log

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/xhrg-product/apollo-client-golang/no_ref"
	"io"
	"time"
)

var ApolloLogger *logrus.Logger

func init() {
	dir := no_ref.HomeDir()
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetReportCaller(true)
	path := dir + "/data/logs/apollo_client_golang/apollo"

	logfile, err := rotatelogs.New(
		path+"_%Y%m%d.log",
		rotatelogs.WithLinkName(path+".log"),
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithMaxAge(3*24*time.Hour),
		rotatelogs.WithRotationTime(time.Hour*24*3),
		rotatelogs.ForceNewFile(),
		rotatelogs.WithClock(rotatelogs.Local),
	)
	if err != nil {
		log.Fatal("error:", err.Error())
	}
	//这里会打印到我的logfile【既/data/logs/apollo_client_golang/apollo下的文件和业务自己配置的logfile
	//既logrus.StandardLogger().Out，当业务不做配置的时候，就是控制台。
	//
	//当业务设置log的级别为error的时候。他的日志里面降不会有apollo的info日志，但是apollo自己的日志文件依然会有info日志。
	//当业务设置log级别为info的时候，他的日志文件和我的日志文件都会有日志。
	log.SetOutput(io.MultiWriter(logfile, logrus.StandardLogger().Out))
	ApolloLogger = log
}

func SetLogLevel(level logrus.Level) {
	ApolloLogger.SetLevel(level)
}

func Log() *logrus.Logger {
	return ApolloLogger
}
