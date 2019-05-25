package master

import (
	"context"
	"fmt"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/utils"
	"w"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
	"time"
)

type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

func InitLogMgr() (err error) {
	var (
		client          *mongo.Client
		configureSource interface{}
	)
	if configureSource, err = common.Manage.GetPrototype("configure"); err != nil {
		fmt.Printf("fail to read file: %v", err)
		return
	}
	configure := configureSource.(utils.IniParser)
	if client, err = mongo.Connect(
		context.TODO(),
		configure.GetString("mongodb", "endpoints"),
		clientopt.ConnectTimeout(time.Duration(configure.GetInt64("mongodb", "connect_timeout"))*time.Millisecond)); err != nil {
		return
	}

	common.Manage.SetSingleton("LogManager", LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	})
	return
}

func (logMgr *LogMgr) ListLog(name string, skip int, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  mongo.Cursor
		jobLog  *common.JobLog
	)

	logArr = make([]*common.JobLog, 0)
	filter = &common.JobLogFilter{JobName: name}
	logSort = &common.SortLogByStartTime{SortOrder: -1}

	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, findopt.Sort(logSort), findopt.Skip(int64(skip)), findopt.Limit(int64(limit))); err != nil {
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}
		if err = cursor.Decode(jobLog); err != nil {
			continue
		}

		logArr = append(logArr, jobLog)
	}
	return
}
