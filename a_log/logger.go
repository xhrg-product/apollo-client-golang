package a_log

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/xhrg-product/apollo-client-golang/no_ref"
	"io"
	"strings"
	"time"
)

var ApolloLogger *MyLog

type MyLog struct {
	*logrus.Logger
}

type MyFormatter struct{}

func (s *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := time.Now().Local().Format("2006-01-02 15:04:05")
	fileName := entry.Caller.File
	fileNames := strings.Split(fileName, "/")
	fileName = fileNames[len(fileNames)-1]
	line := entry.Caller.Line
	msg := fmt.Sprintf("%v [%v] [%v_%v] %v\n", timestamp, strings.ToUpper(entry.Level.String()), fileName, line, entry.Message)
	return []byte(msg), nil
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
	log.SetFormatter(&MyFormatter{})
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
