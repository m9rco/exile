package worker

import (
	"context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/etcd-io/etcd/clientv3"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/utils"
	"strings"
	"time"
)

type JobManager struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

// initialize the job manager
func InitJobMgr() (err error) {
	var (
		config          clientv3.Config
		client          *clientv3.Client
		configureSource interface{}
		jobManageSev    JobManager
	)
	if configureSource, err = common.Manage.GetPrototype("configure"); err != nil {
		return
	}
	configure := configureSource.(utils.IniParser)
	config = clientv3.Config{
		Endpoints:            strings.Split(configure.GetString("etcd", "endpoints"), ","),
		AutoSyncInterval:     0,
		DialTimeout:          time.Duration(configure.GetInt64("etcd", "dial_timeout")) * time.Millisecond,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		Username:             "",
		Password:             "",
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
	}
	if client, err = clientv3.New(config); err != nil {
		return
	}

	common.Manage.SetSingleton("JobManager", JobManager{
		client: client,
		kv:     clientv3.NewKV(client),
		lease:  clientv3.NewLease(client),
		watcher: clientv3.NewWatcher(client),
	})

	jobManageSev = common.Manage.GetSingleton("JobManager").(JobManager)
	if err = jobManageSev.watchJobs(); err != nil {
		return
	}
	jobManageSev.watchKiller()
	return
}

// listen jobs change
func (jobMgr *JobManager) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvPair             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		watchEvent         *clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
		scheduler          Scheduler
	)

	// step 1.  get `/cron/jobs/` all job and get cluster revision
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	scheduler = common.Manage.GetSingleton("Scheduler").(Scheduler)
	for _, kvPair = range getResp.Kvs {
		if job, err = common.UnpackJob(kvPair.Value); err == nil {
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			scheduler.PushJobEvent(jobEvent)
		}
	}
	// step 2. according to this revision listen event change
	go func() {
		watchStartRevision = getResp.Header.Revision + 1
		// listen `/cron/jobs/` change
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE:
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}
				// change push the scheduler
				scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

func (jobMgr *JobManager) watchKiller() {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *common.JobEvent
		jobName    string
		job        *common.Job
		scheduler  Scheduler
	)
	scheduler = common.Manage.GetSingleton("Scheduler").(Scheduler)

	// listen `/cron/killer`
	go func() {
		// listen `/cron/killer/` change
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					// change push the scheduler
					scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:
				}
			}
		}
	}()
}

func (jobMgr *JobManager) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
	return
}
