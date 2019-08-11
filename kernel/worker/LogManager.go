package worker

import (
	"context"
	"fmt"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/utils"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

type LogManager struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

// initialize the InitLogSink
func InitLogSink() (err error) {
	var (
		client          *mongo.Client
		configure       utils.IniParser
		LogManagerSev   LogManager
		configureSource interface{}
	)
	if configureSource, err = common.Manage.GetPrototype("configure"); err != nil {
		fmt.Printf("fail to read file: %v", err)
		return
	}
	configure = configureSource.(utils.IniParser)
	// init a mango connect
	if client, err = mongo.Connect(
		context.TODO(),
		configure.GetString("mongodb", "endpoints"),
		clientopt.ConnectTimeout(time.Duration(configure.GetInt32("mongodb", "connect_timeout"))*time.Millisecond)); err != nil {
		return
	}
	// switch the db and collection
	common.Manage.SetSingleton("LogManager", LogManager{
		client:         client,
		logCollection:  client.Database(configure.GetString("mongodb", "db")).Collection(configure.GetString("mongodb", "logger")),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	})

	LogManagerSev = common.Manage.GetSingleton("LogManager").(LogManager)
	// start mongodb goroutine
	go LogManagerSev.writeLoop()
	return
}

// send logger
func (logSink *LogManager) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
	}
}

// batch write logger
func (logSink *LogManager) saveLogs(batch *common.LogBatch) {
	logSink.logCollection.InsertMany(context.TODO(), batch.Logs)
}

// logger save goroutine
func (logSink *LogManager) writeLoop() {
	var (
		log             *common.JobLog
		logBatch        *common.LogBatch
		commitTimer     *time.Timer
		timeoutBatch    *common.LogBatch
		configure       utils.IniParser
		configureSource interface{}
		err             error
	)

	if configureSource, err = common.Manage.GetPrototype("configure"); err != nil {
		fmt.Printf("fail to read file: %v", err)
		return
	}
	configure = configureSource.(utils.IniParser)

	for {
		select {
		case log = <-logSink.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				// set the timeout automatically submit (1s)
				commitTimer = time.AfterFunc(
					time.Duration(configure.GetInt32("job", "commit_timeout"))*time.Millisecond,
					func(batch *common.LogBatch) func() {
						return func() {
							logSink.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}
			// create new logger append
			logBatch.Logs = append(logBatch.Logs, log)

			// if batch size is full, that is send logger !
			if len(logBatch.Logs) >= int(configure.GetInt32("job", "batch_size")) {
				logSink.saveLogs(logBatch)
				// clean logger batch
				logBatch = nil
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan: // overdue batch
			if timeoutBatch != logBatch {
				continue
			}
			// save loggers
			logSink.saveLogs(timeoutBatch)
			// clean logger batch
			logBatch = nil
		}
	}
}
