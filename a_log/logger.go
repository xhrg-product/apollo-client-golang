package a_log

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/xhrg-product/apollo-client-golang/no_ref"
	"io"
	"time"
)

var ApolloLogger *MyLog

type MyLog struct {
	*logrus.Logger
}

//打印日志到我配置的目录，再打印日志到控制台。
func (mylog *MyLog) Errorf(format string, args ...interface{}) {
	mylog.Logger.Errorf(format, args)
	logrus.Errorf(format, args)
}

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
		//rotatelogs.ForceNewFile(),
		rotatelogs.WithClock(rotatelogs.Local),
	)
	if err != nil {
		log.Fatal("error:", err.Error())
	}
	log.SetOutput(io.MultiWriter(logfile))
	c := &MyLog{Logger: log}
	ApolloLogger = c
}

func SetLogLevel(level logrus.Level) {
	ApolloLogger.SetLevel(level)
}

func Log() *MyLog {
	return ApolloLogger
}
